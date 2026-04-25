package service

import (
	"card-payment-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

func mapRegisterMerchant(data RegisterInput) *domain.Merchant {
	return &domain.Merchant{
		ID:         uuid.New(),
		Name:       data.Name,
		Email:      data.Email,
		Status:     domain.MerchantStatusPending,
		WebhookURL: data.WebhookURL,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func mapRegisterAPIKey(merchantID uuid.UUID, prefix, hashed string) *domain.APIKey {
	return &domain.APIKey{
		ID:         uuid.New(),
		MerchantID: merchantID,
		KeyPrefix:  prefix,
		HashedKey:  hashed,
		CreatedAt:  time.Now(),
	}
}
