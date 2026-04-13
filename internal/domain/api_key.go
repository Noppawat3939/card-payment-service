package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	MerchantID uuid.UUID  `json:"merchant_id"`
	KeyPrefix  string     `json:"key_prefix"`
	HashedKey  string     `json:"-"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
