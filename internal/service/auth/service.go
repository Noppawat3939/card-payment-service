package auth

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/repository"
	"card-payment-service/pkg"
	"context"
)

type AuthService interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (*domain.Merchant, error)
}

type authService struct {
	apiKeyRepo   repository.APIKeyRepository
	merchantRepo repository.MerchantRepository
}

func NewAuthService(
	apiKeyRepo repository.APIKeyRepository,
	merchantRepo repository.MerchantRepository,
) AuthService {
	return &authService{
		apiKeyRepo:   apiKeyRepo,
		merchantRepo: merchantRepo,
	}
}
func (s *authService) ValidateAPIKey(ctx context.Context, apiKey string) (*domain.Merchant, error) {
	hashed := pkg.HashSHA256(apiKey)
	key, err := s.apiKeyRepo.FindByHashedKey(ctx, hashed)
	if err != nil {
		return nil, domain.ErrInvalidApiKey
	}

	// query merchant
	merchant, err := s.merchantRepo.FindByID(ctx, key.MerchantID)
	if err != nil {
		return nil, err
	}
	if merchant.Status != domain.MerchantStatusActive {
		return nil, domain.ErrMerchantNotActive
	}

	return merchant, nil
}
