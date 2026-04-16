package repository

import (
	"card-payment-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdempotencyKeyRepository interface {
	Create(ctx context.Context, data *domain.IdempotencyKey) error
	FindByKeyAndMerchantID(ctx context.Context, key, merchantID uuid.UUID) (*domain.IdempotencyKey, error)
}

type idempotencyKeyRepository struct {
	db *gorm.DB
}

func NewIdempotencyKeyRepository(db *gorm.DB) IdempotencyKeyRepository {
	return &idempotencyKeyRepository{db}
}

func (r *idempotencyKeyRepository) Create(ctx context.Context, data *domain.IdempotencyKey) error {
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *idempotencyKeyRepository) FindByKeyAndMerchantID(ctx context.Context, key, merchantID uuid.UUID) (*domain.IdempotencyKey, error) {
	var data domain.IdempotencyKey

	if err := r.db.WithContext(ctx).
		Where("key = ?", key).
		Where("merchant_id = ?", merchantID).
		Where("expires_at > NOW()").
		First(&data).Error; err != nil {
		return nil, err
	}

	return &data, nil
}
