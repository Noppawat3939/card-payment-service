package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/repository"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PaymentService struct {
	txRepo   repository.TransactionRepository
	idemRepo repository.IdempotencyKeyRepository
	log      zerolog.Logger
	gateway  gateway.Gateway
}

func NewPaymentService(txRepo repository.TransactionRepository, idemRepo repository.IdempotencyKeyRepository, log zerolog.Logger, gateway gateway.Gateway) *PaymentService {
	return &PaymentService{txRepo, idemRepo, log, gateway}
}

type AuthorizeInput struct {
	MerchantID     uuid.UUID
	IdempotencyKey string
	CardNumber     string
	ExpiryMonth    string
	ExpiryYear     string
	CVV            string
	Amount         int64
	Currency       string
	Description    *string
}

type AuthorizeOutput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *PaymentService) Authorize(ctx context.Context, data AuthorizeInput) (*AuthorizeOutput, error) {
	log := s.log.With().
		Str("merchant_id", data.MerchantID.String()).
		Str("idem_key", data.IdempotencyKey).
		Logger()

	// check duplicate idempotency key then return cached
	idemKeyID := uuid.MustParse(data.IdempotencyKey)
	if cached, ok := s.cachedIdempotency(ctx, idemKeyID, data.MerchantID); ok {
		log.Warn().Msg("duplicate idempotency key — returning cached response")
		return cached, nil
	}

	// tokenize card (mock)
	tokenResp, err := s.gateway.TokenizeCard(ctx, gateway.TokenizeRequest{
		CardNumber:  data.CardNumber,
		CVV:         data.CVV,
		ExpiryMonth: data.ExpiryMonth,
		ExpiryYear:  data.ExpiryYear,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to tokenize card")
		return nil, err
	}

	// insert transaction (pending)
	tx := &domain.Transaction{
		ID:             uuid.New(),
		MerchantID:     data.MerchantID,
		PaymentType:    domain.AuthorizeCapture,
		Status:         domain.TransactionStatusPending,
		CardToken:      tokenResp.CardToken,
		CardLastFour:   tokenResp.LastFour,
		CardBrand:      tokenResp.Brand,
		Amount:         data.Amount,
		Currency:       data.Currency,
		Description:    data.Description,
		IdempotencyKey: data.IdempotencyKey,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		log.Error().Err(err).Msg("failed to create transaction")
		return nil, err
	}

	log = log.With().Str("transaction_id", tx.ID.String()).Logger()

	// call gateway
	status, gwRef, reason := s.callGatewayAuthorize(ctx, tx, log)

	// update transaction
	updated, err := s.txRepo.UpdateAndReturn(
		ctx, tx.ID, &domain.Transaction{
			Status:       status,
			FailedReason: reason,
			GatewayRef:   gwRef,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	out := &AuthorizeOutput{
		TransactionID: updated.ID,
		Status:        string(updated.Status),
		CreatedAt:     tx.CreatedAt,
	}

	// save idempotency cache
	s.saveIdempotency(ctx, idemKeyID, data.MerchantID, out, log)

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

func (s *PaymentService) cachedIdempotency(ctx context.Context, key, merchantID uuid.UUID) (*AuthorizeOutput, bool) {
	idem, err := s.idemRepo.FindByKeyAndMerchantID(ctx, key, merchantID)
	if err != nil || idem == nil {
		return nil, false
	}

	var out AuthorizeOutput
	if err := json.Unmarshal(idem.Response, &out); err != nil {
		s.log.Error().Err(err).Msg("failed to unmarshal idempotency response")
		return nil, false
	}

	return &out, true
}

func (s *PaymentService) callGatewayAuthorize(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, gatewayRef, failedReason *string) {
	gw, err := s.gateway.Authorize(ctx, gateway.AuthorizeRequest{
		Amount:   tx.Amount,
		Currency: tx.Currency,
		OrderID:  tx.ID.String(),
	})
	if err != nil || gw == nil {
		log.Error().Err(err).Msg("failed to authorize gateway")
		reason := err.Error()
		return domain.TransactionStatusFailed, nil, &reason
	}

	return domain.TransactionStatusAuthorized, &gw.GatewayRef, nil
}

func (s *PaymentService) saveIdempotency(ctx context.Context, idemKey, merchantID uuid.UUID, out *AuthorizeOutput, log zerolog.Logger) error {
	raw, err := json.Marshal(out)
	if err != nil {
		log.Error().Err(err).Msg("failed to json marshall output")
		return err
	}

	if err = s.idemRepo.Create(ctx, &domain.IdempotencyKey{
		Key:       idemKey,
		MerchatID: merchantID,
		Response:  raw,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 1d
	}); err != nil {
		log.Warn().Err(err).Msg("failed to create idempotency key")
		return err
	}

	return nil
}
