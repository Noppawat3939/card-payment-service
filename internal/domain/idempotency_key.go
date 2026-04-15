package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type IdempotencyKey struct {
	Key       uuid.UUID      `json:"key"`
	MerchatID uuid.UUID      `json:"merchant_id"`
	Response  datatypes.JSON `json:"response"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
}

func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}
