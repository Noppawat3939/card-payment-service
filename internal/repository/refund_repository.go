package repository

import (
	"card-payment-service/internal/domain"
	"context"

	"gorm.io/gorm"
)

type RefundRepository interface {
	Create(ctx context.Context, data *domain.Refund) error
}

type refundRepository struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepository{db}
}

func (r *refundRepository) Create(ctx context.Context, data *domain.Refund) error {
	return r.db.WithContext(ctx).Create(data).Error
}
