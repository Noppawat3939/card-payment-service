package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/repository"
	uuidUtil "card-payment-service/pkg/uuid"
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
	// check duplicate idempotency key then return cached
	idemKey, err := s.idemRepo.FindByKeyAndMerchantID(ctx, *uuidUtil.ParseUUID(data.IdempotencyKey), data.MerchantID)
	if err == nil && idemKey != nil {
		s.log.Warn().Str("merchant_id", data.MerchantID.String()).Str("idem_key", data.IdempotencyKey).Msg("idempotency duplicated")

		var out AuthorizeOutput
		if err := json.Unmarshal(idemKey.Response, &out); err != nil {
			s.log.Error().Err(err).Msg("failed to unmarshal idempotency response")
			return nil, err
		}

		return &out, nil
	}

	// tokenize card (mock)
	cardToken := "tok_" + uuid.NewString()
	lastFour := data.CardNumber[len(data.CardNumber)-4:]

	// insert transaction pending
	tx := &domain.Transaction{
		ID:             uuid.New(),
		MerchantID:     data.MerchantID,
		PaymentType:    domain.AuthorizeCapture,
		Status:         domain.TransactionStatusPending,
		CardToken:      cardToken,
		CardLastFour:   lastFour,
		CardBrand:      "VISA",
		Amount:         data.Amount,
		Currency:       data.Currency,
		Description:    data.Description,
		IdempotencyKey: data.IdempotencyKey,
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		s.log.Error().Err(err).
			Str("merchant_id", data.MerchantID.String()).
			Str("idem_key", data.IdempotencyKey).
			Msg("failed to create a new transaction")
		return nil, err
	}

	// call gateway
	resp, err := s.gateway.Authorize(ctx, gateway.AuthorizeRequest{
		Amount:   tx.Amount,
		Currency: tx.Currency,
		OrderID:  tx.ID.String(),
	})

	var failedReason *string
	updateStatus := domain.TransactionStatusAuthorized

	if err != nil {
		s.log.Error().Err(err).
			Str("transaction_id", tx.ID.String()).
			Msg("failed to call gateway")
		updateStatus = domain.TransactionStatusFailed
		msg := err.Error()
		failedReason = &msg
	}

	if resp == nil {
		s.log.Warn().
			Str("transaction_id", tx.ID.String()).
			Msg("calling to gateway and response is empty")
		updateStatus = domain.TransactionStatusFailed
		failedReason = nil
	}

	// update transaction
	updated, err := s.txRepo.UpdateAndReturn(ctx, tx.ID, &domain.Transaction{Status: updateStatus, FailedReason: failedReason})
	if err != nil {
		s.log.Error().Err(err).Str("transaction_id", tx.ID.String()).Msg("failed to update transaction")
		return nil, err
	}

	// save idempotency cache
	out := &AuthorizeOutput{
		TransactionID: updated.ID,
		Status:        string(updated.Status),
		CreatedAt:     tx.CreatedAt,
	}

	json, err := json.Marshal(out)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to marshall authorize output")
		return nil, err
	}

	err = s.idemRepo.Create(ctx, &domain.IdempotencyKey{
		Key:       *uuidUtil.ParseUUID(data.IdempotencyKey),
		MerchatID: data.MerchantID,
		Response:  json,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 1d
	})
	if err != nil {
		s.log.Error().Err(err).Msg("failed to create a new idempotency key")
		return nil, err
	}

	return out, nil
}
