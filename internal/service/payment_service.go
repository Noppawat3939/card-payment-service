package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PaymentService struct {
	txRepo   repository.TransactionRepository
	idemRepo repository.IdempotencyKeyRepository
	gateway  gateway.Gateway
	log      zerolog.Logger
}

func NewPaymentService(txRepo repository.TransactionRepository, idemRepo repository.IdempotencyKeyRepository, gateway gateway.Gateway, log zerolog.Logger) *PaymentService {
	return &PaymentService{txRepo, idemRepo, gateway, log}
}

type AuthorizeInput struct {
	Amount         int64
	CardNumber     string
	Currency       string
	CVV            string
	Description    *string
	ExpiryMonth    string
	ExpiryYear     string
	IdempotencyKey string
	MerchantID     uuid.UUID
	PaymentType    domain.PaymentType
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
		return nil, domain.ErrTokenizeCard
	}

	// insert transaction (pending)
	tx := &domain.Transaction{
		ID:             uuid.New(),
		MerchantID:     data.MerchantID,
		PaymentType:    data.PaymentType,
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
	query := map[string]interface{}{
		"id":     tx.ID,
		"status": domain.TransactionStatusPending,
	}
	updated, err := s.txRepo.UpdateByQueryAndReturn(
		ctx, query, &domain.Transaction{
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
	if err = s.saveIdempotency(ctx, idemKeyID, data.MerchantID, out, log); err != nil {
		return nil, err
	}

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

type CaptureInput struct {
	TransactionID uuid.UUID
	MerchantID    uuid.UUID
}

type CaptureOutput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
	CapturedAt    time.Time `json:"captured_at"`
}

func (s *PaymentService) Capture(ctx context.Context, data CaptureInput) (*CaptureOutput, error) {
	log := s.log.With().
		Str("merchant_id", data.MerchantID.String()).
		Str("transaction_id", data.TransactionID.String()).
		Logger()

	// check transaction authorized and gateway_ref not null
	tx, err := s.getTxAuthorized(ctx, data.TransactionID, data.MerchantID, log)
	if err != nil {
		return nil, err
	}

	// call gateway
	status, reason := s.callGatewayCapture(ctx, tx, log)

	// update transaction
	query := map[string]interface{}{
		"id":     tx.ID,
		"status": domain.TransactionStatusAuthorized,
	}
	payload := &domain.Transaction{
		Status:       status,
		FailedReason: reason,
	}
	if status == domain.TransactionStatusCaptured {
		now := time.Now()
		payload.CapturedAt = &now
	}

	updated, err := s.txRepo.UpdateByQueryAndReturn(ctx, query, payload)
	if err != nil {
		log.Error().Err(err).Msg("failed update transaction")
		return nil, err
	}

	out := &CaptureOutput{
		TransactionID: updated.ID,
		Status:        string(updated.Status),
		CapturedAt:    *updated.CapturedAt,
	}

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

func (s *PaymentService) cachedIdempotency(ctx context.Context, key, merchantID uuid.UUID) (*AuthorizeOutput, bool) {
	idem, err := s.idemRepo.FindByKeyAndMerchantID(ctx, key, merchantID)
	fmt.Println("idem", idem, "key", key)
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
		Key:        idemKey,
		MerchantID: merchantID,
		Response:   raw,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 1d
	}); err != nil {
		log.Warn().Err(err).Msg("failed to create idempotency key")
		return err
	}

	return nil
}

func (s *PaymentService) getTxAuthorized(ctx context.Context, id, merchantID uuid.UUID, log zerolog.Logger) (*domain.Transaction, error) {
	tx, err := s.txRepo.FindByIDAndMerchantID(ctx, id, merchantID)
	if err != nil {
		log.Error().Err(err).Msg("failed find one transaction")
		return nil, domain.ErrTransactionNotFound
	}
	if tx.GatewayRef == nil || *tx.GatewayRef == "" {
		log.Warn().Msg("invalid gateway_ref")
		return nil, domain.ErrInvalidGatewayRef
	}
	if tx.Status != domain.TransactionStatusAuthorized {
		log.Warn().Msg("status not authorized")
		return nil, domain.ErrTransactionNotCapturable
	}

	return tx, nil
}

func (s *PaymentService) callGatewayCapture(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, failedReason *string) {
	resp, err := s.gateway.Capture(ctx, gateway.CaptureRequest{
		OrderID:    tx.ID.String(),
		GatewayRef: *tx.GatewayRef,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to capture gateway")
		reason := err.Error()
		return domain.TransactionStatusFailed, &reason
	}
	if resp == nil {
		log.Error().Err(err).Msg("gateway returned empty response")
		reason := "empty gateway response"
		return domain.TransactionStatusFailed, &reason
	}

	return domain.TransactionStatusCaptured, nil
}
