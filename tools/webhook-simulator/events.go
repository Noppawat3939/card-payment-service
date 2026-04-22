package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type EventType string
type WebhookStatus string

const (
	EventRefundCompleted EventType = "refund.completed"
	EventRefundFailed    EventType = "refund.failed"
	EventPaymentSuccess  EventType = "payment.success"
	EventPaymentFailed   EventType = "payment.failed"
)

const (
	WebhookStatusCompleted WebhookStatus = "completed"
	WebhookStatusFailed    WebhookStatus = "failed"
	WebhookStatusCaptured  WebhookStatus = "captured"
)

type WebhookPayload struct {
	Event      EventType     `json:"event"`
	RefundRef  string        `json:"refund_ref,omitempty"`
	GatewayRef string        `json:"gateway_ref"`
	Status     WebhookStatus `json:"status"`
	Amount     int64         `json:"amount,omitempty"`
	Currency   string        `json:"currency"`
	OccurredAt time.Time     `json:"occurred_at"`
}

func buildPayload(event EventType, gatewayRef string) (*WebhookPayload, error) {
	base := &WebhookPayload{
		Event:      event,
		GatewayRef: gatewayRef,
		Currency:   "THB",
		OccurredAt: time.Now(),
	}

	switch event {
	case EventRefundCompleted:
		base.RefundRef = "rf_" + uuid.NewString()
		base.Status = WebhookStatusCompleted
		base.Amount = 100000

	case EventRefundFailed:
		base.RefundRef = "rf_" + uuid.NewString()
		base.Status = WebhookStatusFailed
		base.Amount = 100000

	case EventPaymentSuccess:
		base.Status = WebhookStatusCaptured
		base.Amount = 150000

	case EventPaymentFailed:
		base.Status = WebhookStatusFailed

	default:
		return nil, fmt.Errorf("unknown event type: %s", event)

	}

	return base, nil
}

func availableEvents() []string {
	return []string{
		string(EventRefundCompleted),
		string(EventRefundFailed),
		string(EventPaymentSuccess),
		string(EventPaymentFailed),
	}
}
