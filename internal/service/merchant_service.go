package service

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/repository"
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type MerchantService struct {
	merchantRepo repository.MerchantRepository
	apiKeyRepo   repository.APIKeyRepository
	log          zerolog.Logger
}

func NewMerchantService(merchantRepo repository.MerchantRepository, apiKeyRepo repository.APIKeyRepository, log zerolog.Logger) *MerchantService {
	return &MerchantService{merchantRepo, apiKeyRepo, log}
}

type RegisterInput struct {
	Name       string
	Email      string
	WebhookURL *string
}

type RegisterOutput struct {
	MerchantID string `json:"merchant_id"`
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	Status     string `json:"status"`
}

func (s *MerchantService) Register(ctx context.Context, data RegisterInput) (*RegisterOutput, error) {
	// check existing
	existing, err := s.merchantRepo.FindByEmail(ctx, data.Email)
	if err != nil {
		s.log.Error().Str("email", data.Email).Msg("failed to find one merchant")
		return nil, err
	}
	if existing != nil {
		s.log.Warn().Str("email", data.Email).Msg("merchant already exists")
		return nil, domain.ErrMerchantAlreadyExists
	}

	// create merchant
	merchant := mapRegisterMerchant(data)

	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		s.log.Error().Msg("failed to create a new merchant")
		return nil, err
	}

	// generate api-key, secret and create
	secret := generateSecret()
	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error().Msg("failed to hash secret for register merchant")
		return nil, err
	}

	apiKey := mapRegisterAPIKey(merchant, secret[:16], string(hashed))

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		s.log.Error().Msg("failed to create a new api-key")
		return nil, err
	}

	return &RegisterOutput{
		MerchantID: merchant.ID.String(),
		APIKey:     apiKey.HashedKey,
		APISecret:  apiKey.KeyPrefix,
		Status:     string(merchant.Status),
	}, nil
}

func (s *MerchantService) Activate(ctx context.Context, email string) (*domain.Merchant, error) {
	// check existing and not pending status
	existings, err := s.merchantRepo.FindByEmail(ctx, email)
	if err != nil {
		s.log.Error().Str("email", email).Msg("failed to find merchant by email")
		return nil, err
	}
	if existings == nil {
		s.log.Warn().Str("email", email).Msg("merchant not found")
		return nil, domain.ErrMerchantNotFound
	}
	if existings.Status != domain.MerchantStatusPending {
		s.log.Warn().Str("email", email).Msg("current status not active")
		return nil, domain.ErrMerchantStatusNotAccepted
	}

	// update status to active
	values := map[string]interface{}{"status": domain.MerchantStatusActive}
	merchant, err := s.merchantRepo.UpdateAndReturn(ctx, existings.ID, values)
	if err != nil {
		s.log.Error().Str("email", email).Msg("failed to update merchant by email")
		return nil, err
	}

	return merchant, nil
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
