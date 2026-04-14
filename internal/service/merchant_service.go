package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/repository"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	WebhookURL *string `json:"webhook_url"`
}

type RegisterOutput struct {
	MerchantID string `json:"merchant_id"`
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	Status     string `json:"status"`
}

type MerchantService struct {
	merchantRepo repository.MerchantRepository
	apiKeyRepo   repository.APIKeyRepository
}

func NewMerchantService(merchantRepo repository.MerchantRepository, apiKeyRepo repository.APIKeyRepository) *MerchantService {
	return &MerchantService{merchantRepo, apiKeyRepo}
}

func (s *MerchantService) Register(ctx context.Context, data RegisterInput) (*RegisterOutput, error) {
	// check existing
	existing, _ := s.merchantRepo.FindByEmail(ctx, data.Email)
	if existing != nil {
		return nil, domain.ErrMerchantAlreadyExists
	}

	// create merchant
	merchant := mapRegisterMerchant(data)

	if e := s.merchantRepo.Create(ctx, merchant); e != nil {
		return nil, e
	}

	// generate api-key, secret and create
	secret := generateSecret()
	hashed, e := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if e != nil {
		return nil, e
	}

	apiKey := mapRegisterAPIKey(merchant, secret[:16], string(hashed))

	if e := s.apiKeyRepo.Create(ctx, apiKey); e != nil {
		return nil, e
	}

	return &RegisterOutput{
		MerchantID: merchant.ID.String(),
		APIKey:     apiKey.HashedKey,
		APISecret:  apiKey.KeyPrefix,
		Status:     string(merchant.Status),
	}, nil
}

func generateSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "sk_live_" + hex.EncodeToString(b)
}

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

func mapRegisterAPIKey(merchant *domain.Merchant, secret, hashed string) *domain.APIKey {
	return &domain.APIKey{
		ID:         uuid.New(),
		MerchantID: merchant.ID,
		KeyPrefix:  secret,
		HashedKey:  hashed,
		CreatedAt:  time.Now(),
	}
}
