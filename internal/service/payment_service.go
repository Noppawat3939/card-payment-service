package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/infra/redis"
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
	locker   redis.Locker
	log      zerolog.Logger
}

func NewPaymentService(txRepo repository.TransactionRepository,
	idemRepo repository.IdempotencyKeyRepository,
	gateway gateway.Gateway,
	locker redis.Locker,
	log zerolog.Logger,
) *PaymentService {
	return &PaymentService{
		txRepo,
		idemRepo,
		gateway,
		locker,
		log,
	}
}

type ChargeInput struct {
	Amount         int64
	CardNumber     string
	Currency       string
	CVV            string
	Description    *string
	ExpiryMonth    string
	ExpiryYear     string
	IdempotencyKey string
	MerchantID     uuid.UUID
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

type ChargeOutput struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Status        string    `json:"status"`
}

func (s *PaymentService) Authorize(ctx context.Context, data ChargeInput) (*ChargeOutput, error) {
	// initialize log Authorize service
	log := s.log.With().
		Str("func", "Authorize").
		Str("merchant_id", data.MerchantID.String()).
		Str("idem_key", data.IdempotencyKey).
		Logger()

	// check duplicate idempotency key then return cached
	idemKeyID := uuid.MustParse(data.IdempotencyKey)
	if cached, ok := s.cachedIdempotency(ctx, idemKeyID, data.MerchantID, log); ok {
		return cached, nil
	}

	// tokenize card (mock)
	tokenizedResp, err := s.tokenizeCardToGateway(ctx, data, log)
	if err != nil {
		return nil, err
	}

	// save transaction (pending)
	tx, err := s.saveTransaction(ctx, data, domain.AuthorizeCapture, tokenizedResp, log)
	if err != nil {
		return nil, err
	}

	// re-assign log to transaction_id
	log = log.With().
		Str("func", "Authorize").
		Str("transaction_id", tx.ID.String()).
		Logger()

	// call gateway (authorize)
	status, gwRef, reason := s.callGatewayAuthorize(ctx, tx, log)

	// update transaction
	queryUpdate := map[string]any{
		"id":     tx.ID,
		"status": domain.TransactionStatusPending,
	}
	updatePayload := domain.Transaction{Status: status}

	if status == domain.TransactionStatusFailed {
		updatePayload.FailedReason = reason
	} else {
		updatePayload.GatewayRef = gwRef
	}

	updated, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed update transaction")
		return nil, err
	}

	out := &ChargeOutput{
		TransactionID: updated.ID,
		Status:        string(updated.Status),
	}

	// save idempotency + response cache
	if err = s.saveIdempotency(ctx, idemKeyID, data.MerchantID, out, log); err != nil {
		return nil, err
	}

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

func (s *PaymentService) Capture(ctx context.Context, data CaptureInput) (*CaptureOutput, error) {
	// initialize log Capture service
	log := s.log.With().
		Str("func", "Capture").
		Str("merchant_id", data.MerchantID.String()).
		Str("transaction_id", data.TransactionID.String()).
		Logger()

	// lock transaction for update
	lockValue, lockKey, err := s.aquireLock(ctx, data.TransactionID, log)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = s.locker.Release(context.Background(), *lockKey, *lockValue)
	}()

	// check transaction authorized and gateway_ref not null
	tx, err := s.getTxAuthorized(ctx, data.TransactionID, data.MerchantID, log)
	if err != nil {
		return nil, err
	}

	// call gateway
	status, _, reason := s.callGatewayCapture(ctx, tx, log)

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

func (s *PaymentService) Charge(ctx context.Context, data ChargeInput) (*ChargeOutput, error) {
	// initialize log Charge service
	log := s.log.With().
		Str("func", "Charge").
		Str("merchant_id", data.MerchantID.String()).
		Str("idem_key", data.IdempotencyKey).
		Logger()

	// check duplicate idempotency key then return cached
	idemKeyID := uuid.MustParse(data.IdempotencyKey)
	if cached, ok := s.cachedIdempotency(ctx, idemKeyID, data.MerchantID, log); ok {
		return cached, nil
	}

	// tokenize card (mock)
	tokenizedResp, err := s.tokenizeCardToGateway(ctx, data, log)
	if err != nil {
		return nil, err
	}

	// save transaction (pending)
	tx, err := s.saveTransaction(ctx, data, domain.DirectCharge, tokenizedResp, log)
	if err != nil {
		return nil, err
	}

	// re-assign log to transaction_id
	log = log.With().
		Str("func", "Charge").
		Str("transaction_id", tx.ID.String()).
		Logger()

	// call gateway (capture)
	status, gwRef, reason := s.callGatewayCapture(ctx, tx, log)

	// update transaction (captured/failed)
	queryUpdate := map[string]any{
		"id":     tx.ID,
		"status": domain.TransactionStatusPending,
	}
	updatePayload := &domain.Transaction{Status: status}

	if status == domain.TransactionStatusFailed {
		updatePayload.FailedReason = reason
	} else {
		updatePayload.GatewayRef = gwRef
	}

	updated, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed update transaction")
		return nil, err
	}

	out := &ChargeOutput{TransactionID: updated.ID, Status: string(updated.Status)}

	// save idempotency + response cache
	if err := s.saveIdempotency(ctx, idemKeyID, data.MerchantID, out, log); err != nil {
		return nil, err
	}

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

func (s *PaymentService) cachedIdempotency(ctx context.Context, key, merchantID uuid.UUID, log zerolog.Logger) (*ChargeOutput, bool) {
	idem, err := s.idemRepo.FindByKeyAndMerchantID(ctx, key, merchantID)

	if err != nil || idem == nil {
		log.Warn().Msg("duplicate idempotency key — returning cached response")
		return nil, false
	}

	var out ChargeOutput
	if err := json.Unmarshal(idem.Response, &out); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal idempotency response")
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
		log.Error().Err(err).Msg("failed authorize gateway")
		reason := err.Error()
		return domain.TransactionStatusFailed, nil, &reason
	}

	return domain.TransactionStatusAuthorized, &gw.GatewayRef, nil
}

func (s *PaymentService) saveIdempotency(ctx context.Context, idemKey, merchantID uuid.UUID, out *ChargeOutput, log zerolog.Logger) error {
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

func (s *PaymentService) tokenizeCardToGateway(ctx context.Context, data ChargeInput, log zerolog.Logger) (*gateway.TokenizeResponse, error) {
	req := gateway.TokenizeRequest{
		CardNumber:  data.CardNumber,
		CVV:         data.CVV,
		ExpiryMonth: data.ExpiryMonth,
		ExpiryYear:  data.ExpiryYear,
	}

	tokenizeResp, err := s.gateway.TokenizeCard(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("failed tokenize card")
		return nil, domain.ErrTokenizeCard
	}

	return tokenizeResp, nil
}

func (s *PaymentService) saveTransaction(ctx context.Context,
	data ChargeInput,
	paymentType domain.PaymentType,
	tokenized *gateway.TokenizeResponse,
	log zerolog.Logger) (*domain.Transaction, error) {
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
		log.Error().Err(err).Msgf("failed create transaction type %s", paymentType)
		return nil, err
	}

	return tx, nil
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

func (s *PaymentService) callGatewayCapture(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, gatewayRef, failedReason *string) {
	resp, err := s.gateway.Capture(ctx, gateway.CaptureRequest{
		OrderID:    tx.ID.String(),
		GatewayRef: *tx.GatewayRef,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to capture gateway")
		reason := err.Error()
		return domain.TransactionStatusFailed, nil, &reason
	}
	if resp == nil {
		log.Error().Err(err).Msg("gateway returned empty response")
		reason := "empty gateway response"
		return domain.TransactionStatusFailed, nil, &reason
	}

	return domain.TransactionStatusCaptured, &resp.GatewayRef, nil
}

func (s *PaymentService) aquireLock(ctx context.Context, transactionID uuid.UUID, log zerolog.Logger) (*string, *string, error) {
	lockKey := fmt.Sprintf("lock:tx:%s", transactionID)

	lockValue, err := s.locker.Acquire(ctx, lockKey, 5*time.Second)
	if err != nil {
		s.log.Error().Err(err).Msg("failed acquire lock")
		return nil, nil, err
	}
	if lockKey == "" {
		s.log.Warn().Msg("duplicated request")
		return nil, nil, domain.ErrDuplicateRequest
	}

	// lock success
	return &lockKey, &lockValue, nil
}
