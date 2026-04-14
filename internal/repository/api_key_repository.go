package repository

import (
	"card-payment-service/internal/domain"
	"context"

	"gorm.io/gorm"
)

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *domain.APIKey) error
	FindByHashedKey(ctx context.Context, hashedKey string) (*domain.APIKey, error)
}

type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) APIKeyRepository {
	return &apiKeyRepository{db}
}

func (r *apiKeyRepository) Create(ctx context.Context, data *domain.APIKey) error {

	return r.db.WithContext(ctx).Create(data).Error
}

func (r *apiKeyRepository) FindByHashedKey(ctx context.Context, hashedKey string) (*domain.APIKey, error) {
	var data domain.APIKey
	if e := r.db.WithContext(ctx).Where("hashed_key = ?", hashedKey).First(&data).Error; e != nil {
		return nil, e
	}

	return &data, nil
}
