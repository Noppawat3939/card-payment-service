package repository

import (
	"card-payment-service/internal/domain"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MerchantRepository interface {
	Create(ctx context.Context, data *domain.Merchant) error
	FindByEmail(ctx context.Context, email string) (*domain.Merchant, error)
	UpdateAndReturn(ctx context.Context, merchantID uuid.UUID, data interface{}) (*domain.Merchant, error)
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

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *merchangeRepository) UpdateAndReturn(ctx context.Context, merchantID uuid.UUID, values interface{}) (*domain.Merchant, error) {
	var data domain.Merchant
	if err := r.db.WithContext(ctx).
		Model(&data).
		Clauses(clause.Returning{}).
		Where("id = ?", merchantID).
		Updates(values).Error; err != nil {
		return nil, err
	}

	return &data, nil
}
