package domain

import (
	"time"

	"github.com/google/uuid"
)

type MerchantStatus string

const (
	MerchantStatusPending   MerchantStatus = "pending"
	MerchantStatusActive    MerchantStatus = "active"
	MerchantStatusSuspended MerchantStatus = "suspended"
)

type Merchant struct {
	ID         uuid.UUID      `json:"id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	WebhookURL *string        `json:"webhook_url,omitempty"`
	Status     MerchantStatus `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}
