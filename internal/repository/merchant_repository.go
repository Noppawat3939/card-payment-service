package repository

import (
	"card-payment-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantRepository interface {
	Create(ctx context.Context, data *domain.Merchant) error
	FindByEmail(ctx context.Context, email string) (*domain.Merchant, error)
	UpdateStatus(ctx context.Context, merchantID uuid.UUID, status domain.MerchantStatus) error
}

type merchangeRepository struct {
	db *gorm.DB
}

func NewMerchantRepository(db *gorm.DB) MerchantRepository {
	return &merchangeRepository{db}
}

func (r *merchangeRepository) Create(ctx context.Context, data *domain.Merchant) error {
	return r.db.WithContext(ctx).Create(data).Error
}

func (r *merchangeRepository) FindByEmail(ctx context.Context, email string) (*domain.Merchant, error) {
	var data domain.Merchant

	if e := r.db.WithContext(ctx).Where("email = ?", email).First(&data).Error; e != nil {
		return nil, e
	}
	return &data, nil
}

func (r *merchangeRepository) UpdateStatus(ctx context.Context, merchantID uuid.UUID, status domain.MerchantStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Merchant{}).Where("id = ?", merchantID).Update("status", status).Error
}
