package payment

// expose

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"card-payment-service/internal/infra/redis"
	"card-payment-service/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type PaymentService struct {
	txRepo     repository.TransactionRepository
	idemRepo   repository.IdempotencyKeyRepository
	refundRepo repository.RefundRepository
	gateway    gateway.Gateway
	locker     redis.Locker
	log        zerolog.Logger
}

func NewPaymentService(txRepo repository.TransactionRepository,
	idemRepo repository.IdempotencyKeyRepository,
	refundRepo repository.RefundRepository,
	gateway gateway.Gateway,
	locker redis.Locker,
	log zerolog.Logger,
) *PaymentService {
	return &PaymentService{
		txRepo,
		idemRepo,
		refundRepo,
		gateway,
		locker,
		log,
	}
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
	var out ChargeOutput
	if ok := s.cachedIdempotency(ctx, idemKeyID, data.MerchantID, &out, log); ok {
		return &out, nil
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

	// re-assign child
	log = log.With().Str("func", "Authorize").Str("transaction_id", tx.ID.String()).Logger()

	// call gateway (authorize)
	status, gwRef, reason := s.callGatewayAuthorize(ctx, tx, log)

	// update transaction
	queryUpdate := map[string]any{"id": tx.ID, "status": domain.TransactionStatusPending}
	updatePayload := domain.Transaction{Status: status}
	if status == domain.TransactionStatusFailed {
		updatePayload.FailedReason = reason
	} else {
		updatePayload.GatewayRef = gwRef
	}

	updatedResp, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	out = ChargeOutput{
		TransactionID: updatedResp.ID,
		Status:        updatedResp.Status,
	}

	// save idempotency in backgroud
	go s.saveIdempotency(idemKeyID, data.MerchantID, &out, log)

	if status == domain.TransactionStatusFailed {
		return &out, domain.ErrGatewayRejected
	}

	return &out, nil
}

func (s *PaymentService) Capture(ctx context.Context, data CaptureInput) (*CaptureOutput, error) {
	// initialize log Capture service
	log := s.log.With().
		Str("func", "Capture").
		Str("merchant_id", data.MerchantID.String()).
		Str("transaction_id", data.TransactionID.String()).
		Logger()

	// lock transaction for update
	lockKey, lockValue, err := s.aquireLock(ctx, data.TransactionID, log)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := s.locker.Release(context.Background(), *lockKey, *lockValue); err != nil {
			log.Warn().Err(err).Msg("failed to release locker")
		}
	}()

	// check transaction authorized and gateway_ref not null
	tx, err := s.getTxAuthorized(ctx, data.TransactionID, data.MerchantID, log)
	if err != nil {
		return nil, err
	}

	// call gateway
	status, _, reason := s.callGatewayCapture(ctx, tx, log)

	// update transaction
	queryUpdate := map[string]interface{}{"id": tx.ID, "status": domain.TransactionStatusAuthorized}
	updatePayload := &domain.Transaction{Status: status, FailedReason: reason}
	if status == domain.TransactionStatusCaptured {
		now := time.Now()
		updatePayload.CapturedAt = &now
	}

	updatedResp, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	out := &CaptureOutput{
		TransactionID: updatedResp.ID,
		Status:        updatedResp.Status,
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
	var out ChargeOutput
	if ok := s.cachedIdempotency(ctx, idemKeyID, data.MerchantID, &out, log); ok {
		return &out, nil
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

	// re-assign child
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

	updatedResp, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	out = ChargeOutput{TransactionID: updatedResp.ID, Status: updatedResp.Status}

	// save idempotency + response cache
	go s.saveIdempotency(idemKeyID, data.MerchantID, &out, log)

	if status == domain.TransactionStatusFailed {
		return &out, domain.ErrGatewayRejected
	}

	return &out, nil
}

func (s *PaymentService) Void(ctx context.Context, data VoidInput) (*VoidOutput, error) {
	// intialize log Void service
	log := s.log.With().
		Str("func", "Void").
		Str("merchant_id", data.MerchantID.String()).
		Str("transaction_id", data.TransactionID.String()).
		Logger()

	// lock transaction for update
	lockKey, lockValue, err := s.aquireLock(ctx, data.TransactionID, log)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := s.locker.Release(context.Background(), *lockKey, *lockValue); err != nil {
			log.Warn().Err(err).Msg("failed to release locker")
		}
	}()

	// query transaction authorized
	tx, err := s.getTxAuthorized(ctx, data.TransactionID, data.MerchantID, log)
	if err != nil {
		return nil, err
	}
	if tx.Status == domain.TransactionStatusVoided {
		return nil, domain.ErrTransactionAlreadyVoided
	}

	// call gateway
	status, reason := s.callGatewayVoid(ctx, tx, log)

	// update transaction
	queryUpdate := map[string]interface{}{"id": tx.ID, "status": domain.TransactionStatusAuthorized}
	updatePayload := &domain.Transaction{Status: status, FailedReason: reason}

	updatedResp, err := s.txRepo.UpdateByQueryAndReturn(ctx, queryUpdate, updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("failed to update transaction")
		return nil, err
	}

	out := &VoidOutput{
		TransactionID: updatedResp.ID,
		Status:        updatedResp.Status,
	}

	if status == domain.TransactionStatusFailed {
		return out, domain.ErrGatewayRejected
	}

	return out, nil
}

func (s *PaymentService) Refund(ctx context.Context, data RefundInput) (*RefundOutput, error) {
	// initialize log Refund service
	log := s.log.With().
		Str("func", "Refund").
		Str("transaction_id", data.TransactionID.String()).
		Str("merchan_id", data.MerchantID.String()).
		Logger()

	// check duplicate idempotency key then return cached
	var out RefundOutput
	if ok := s.cachedIdempotency(ctx, data.IdempotencyKey, data.MerchantID, &out, log); ok {
		return &out, nil
	}

	// lock transaction for update
	lockKey, lockValue, err := s.aquireLock(ctx, data.TransactionID, log)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := s.locker.Release(context.Background(), *lockKey, *lockValue); err != nil {
			log.Warn().Err(err).Msg("failed to release locker")
		}
	}()

	// check transaction captured
	tx, err := s.txRepo.FindByIDAndMerchantID(ctx, data.TransactionID, data.MerchantID)
	if err != nil {
		return nil, err
	}
	if tx.Status == domain.TransactionStatusRefunded {
		return nil, domain.ErrTransactionAlreadyRefunded
	}
	if tx.Status != domain.TransactionStatusCaptured {
		return nil, domain.ErrTransactionNotRefundable
	}

	// call gateway
	refundRef, _, err := s.callGatewayRefund(ctx, tx, log)
	if err != nil {
		return nil, err
	}

	// insert refund (processing) and update transaction
	refund := &domain.Refund{
		ID:            uuid.New(),
		TransactionID: tx.ID,
		MerchantID:    data.MerchantID,
		Amount:        tx.Amount,
		Status:        domain.RefundProcessing,
		RefundRef:     refundRef,
	}
	if err := s.createRefundAndUpdateTx(ctx, tx.ID, refund, log); err != nil {
		return nil, err
	}

	out = RefundOutput{
		RefundID: refund.ID,
		Status:   refund.Status,
	}
	// save idempotency in backgroud
	go s.saveIdempotency(data.IdempotencyKey, data.MerchantID, &out, log)

	return &out, nil
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
		log.Error().Err(err).Msg("failed to tokenize card")
		return nil, domain.ErrTokenizeCard
	}

	return tokenizeResp, nil
}

func (s *PaymentService) aquireLock(ctx context.Context, transactionID uuid.UUID, log zerolog.Logger) (*string, *string, error) {
	lockKey := fmt.Sprintf("lock:tx:%s", transactionID)

	lockValue, err := s.locker.Acquire(ctx, lockKey, 5*time.Second)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to acquire lock")
		return nil, nil, err
	}
	if lockValue == "" {
		s.log.Warn().Msg("duplicated request")
		return nil, nil, domain.ErrDuplicateRequest
	}

	// lock success
	return &lockKey, &lockValue, nil
}
