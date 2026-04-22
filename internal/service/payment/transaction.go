package payment

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func (s *PaymentService) getTxAuthorized(ctx context.Context, id, merchantID uuid.UUID, log zerolog.Logger) (*domain.Transaction, error) {
	tx, err := s.txRepo.FindByIDAndMerchantID(ctx, id, merchantID)
	if err != nil {
		log.Error().Err(err).Msg("failed to find transaction")
		return nil, domain.ErrTransactionNotFound
	}
	if tx.GatewayRef == nil {
		log.Warn().Msg("invalid gateway_ref")
		return nil, domain.ErrInvalidGatewayRef
	}
	if tx.Status != domain.TransactionStatusAuthorized {
		log.Warn().Msg("status not authorized")
		return nil, domain.ErrTransactionNotCapturable
	}

	return tx, nil
}

func (s *PaymentService) createRefundAndUpdateTx(ctx context.Context, txID uuid.UUID, refund *domain.Refund, log zerolog.Logger) error {
	queryUpdate := map[string]interface{}{
		"id":     txID,
		"status": domain.TransactionStatusCaptured,
	}
	payload := map[string]interface{}{
		"status": domain.TransactionStatusRefunded,
	}
	_, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, payload)
	if err != nil {
		log.Warn().Err(err).Msg("failed to update transaction status to refunded")
		return err
	}

	// insert refund (processing)
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		log.Error().Err(err).Msg("failed to insert refund")
		return err
	}

	return nil
}

func (s *PaymentService) saveTransaction(ctx context.Context,
	data ChargeInput,
	paymentType domain.PaymentType,
	tokenized *gateway.TokenizeResponse,
	log zerolog.Logger) (*domain.Transaction, error) {
	// build payload
	tx := &domain.Transaction{
		Amount:         data.Amount,
		CardBrand:      tokenized.Brand,
		CardLastFour:   tokenized.LastFour,
		CardToken:      tokenized.CardToken,
		Currency:       data.Currency,
		Description:    data.Description,
		ID:             uuid.New(),
		IdempotencyKey: data.IdempotencyKey,
		MerchantID:     data.MerchantID,
		PaymentType:    paymentType,
		Status:         domain.TransactionStatusPending, // initial
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		log.Error().Err(err).Msgf("failed to save transaction with payment type: %s", paymentType)
		return nil, err
	}

	return tx, nil
}
