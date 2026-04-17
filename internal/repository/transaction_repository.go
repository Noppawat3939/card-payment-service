package repository

import (
	"card-payment-service/internal/domain"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TransactionRepository interface {
	Create(ctx context.Context, data *domain.Transaction) error
	FindByIDAndMerchantID(ctx context.Context, id, merchantID uuid.UUID) (*domain.Transaction, error)
	UpdateByQueryAndReturn(ctx context.Context, query map[string]interface{}, values interface{}) (*domain.Transaction, error)
}

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db}
}

func (r *transactionRepository) Create(ctx context.Context, data *domain.Transaction) error {
	err := r.db.WithContext(ctx).Create(data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDuplicateIdempotencyKey
		}
		return err
	}

	return nil
}

func (r *transactionRepository) FindByIDAndMerchantID(ctx context.Context, id, merchantID uuid.UUID) (*domain.Transaction, error) {
	var data domain.Transaction

	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Where("merchant_id = ?", merchantID).
		First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *transactionRepository) UpdateByQueryAndReturn(ctx context.Context, query map[string]interface{}, values interface{}) (*domain.Transaction, error) {
	var data domain.Transaction
	if err := r.db.WithContext(ctx).
		Model(&data).Clauses(clause.Returning{}).
		Where(query).
		Updates(values).Error; err != nil {
		return nil, err
	}

	return &data, nil
}
