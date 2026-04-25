package httpx

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/middleware"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentRequestMeta struct {
	TransactionID  uuid.UUID
	MerchantID     uuid.UUID
	IdempotencyKey uuid.UUID
}

func GetTransactionID(c *gin.Context) (uuid.UUID, error) {
	txStr := c.Param("transaction_id")

	if txStr == "" {
		return uuid.Nil, domain.ErrMissingTransactionID
	}

	txID, err := uuid.Parse(txStr)

	if err != nil {
		return uuid.Nil, err
	}

	return txID, nil
}

func GetMerchantID(c *gin.Context) (uuid.UUID, error) {
	val, ok := c.Get(middleware.MerchantIDKey)
	if !ok {
		return uuid.Nil, domain.ErrMissingMerchantID
	}

	merchantID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("merchant_id invalid type")
	}

	return merchantID, nil
}

func GetIdempotencyKey(c *gin.Context) (uuid.UUID, error) {
	val, ok := c.Get(middleware.IdempotencyKeyContextKey)
	if !ok {
		return uuid.Nil, errors.New("missing idempotency key")
	}

	idemKey, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("idempotency key invalid type")
	}

	return idemKey, nil
}

func GetPaymentMeta(c *gin.Context) (*PaymentRequestMeta, error) {
	txID, err := GetTransactionID(c)
	if err != nil {
		return nil, err
	}

	merchantID, err := GetMerchantID(c)
	if err != nil {
		return nil, err
	}

	idemKey, err := GetIdempotencyKey(c)
	if err != nil {
		return nil, err
	}

	return &PaymentRequestMeta{
		TransactionID:  txID,
		MerchantID:     merchantID,
		IdempotencyKey: idemKey,
	}, nil
}
