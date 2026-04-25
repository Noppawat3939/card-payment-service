package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/gateway"
	"context"

	"github.com/rs/zerolog"
)

func (s *PaymentService) callGatewayAuthorize(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, gatewayRef, failedReason *string) {
	var reason string

	resp, err := s.gateway.Authorize(ctx, gateway.AuthorizeRequest{
		Amount:   tx.Amount,
		Currency: tx.Currency,
		OrderID:  tx.ID.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to authorize gateway")
		reason = err.Error()
		return domain.TransactionStatusFailed, nil, &reason
	}
	if resp == nil {
		log.Warn().Msg("gateway authorized empty response")
		reason = "empty gateway response"
		return domain.TransactionStatusFailed, nil, &reason
	}

	return domain.TransactionStatusAuthorized, &resp.GatewayRef, nil
}

func (s *PaymentService) callGatewayCapture(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, gatewayRef, failedReason *string) {
	var reason string

	resp, err := s.gateway.Capture(ctx, gateway.CaptureRequest{
		OrderID:    tx.ID.String(),
		GatewayRef: *tx.GatewayRef,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to capture gateway")
		reason = err.Error()
		return domain.TransactionStatusFailed, nil, &reason
	}
	if resp == nil {
		log.Warn().Msg("gateway captured empty response")
		reason = "empty gateway response"
		return domain.TransactionStatusFailed, nil, &reason
	}

	return domain.TransactionStatusCaptured, &resp.GatewayRef, nil
}

func (s *PaymentService) callGatewayVoid(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (status domain.TransactionStatus, failedReason *string) {
	var reason string

	resp, err := s.gateway.Void(ctx, gateway.VoidRequest{GatewayRef: *tx.GatewayRef, OrderID: tx.ID.String()})
	if err != nil {
		log.Error().Err(err).Msg("failed to void gateway")
		reason = err.Error()
		return domain.TransactionStatusFailed, &reason
	}
	if resp == nil {
		log.Warn().Msg("gateway returned empty response")
		reason = "empty gateway response"
		return domain.TransactionStatusFailed, &reason
	}
	return domain.TransactionStatusVoided, nil
}

func (s *PaymentService) callGatewayRefund(ctx context.Context, tx *domain.Transaction, log zerolog.Logger) (refundRef, failedReason *string, err error) {
	var reason string

	resp, err := s.gateway.Refund(ctx, gateway.RefundRequest{GatewayRef: *tx.GatewayRef, Amount: tx.Amount})
	if err != nil {
		log.Error().Err(err).Msg("failed to refund gateway")
		reason = err.Error()
		return nil, &reason, err
	}
	if resp == nil {
		log.Warn().Msg("gateway returned empty response")
		reason = "empty gateway response"
		return nil, &reason, domain.ErrGatewayRejected
	}

	return &resp.RefundRef, nil, nil
}
