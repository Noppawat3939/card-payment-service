package merchant

import (
	"card-payment-service/internal/domain"
	"card-payment-service/internal/repository"
	"context"

	"github.com/rs/zerolog"
)

// expose
type MerchantService struct {
	merchantRepo repository.MerchantRepository
	apiKeyRepo   repository.APIKeyRepository
	log          zerolog.Logger
}

func NewMerchantService(merchantRepo repository.MerchantRepository, apiKeyRepo repository.APIKeyRepository, log zerolog.Logger) *MerchantService {
	return &MerchantService{merchantRepo, apiKeyRepo, log}
}

func (s *MerchantService) Register(ctx context.Context, data RegisterInput) (*RegisterOutput, error) {
	log := s.log.With().Str("func", "Register").Str("email", data.Email).Logger()

	// check existing
	existing, err := s.merchantRepo.FindByEmail(ctx, data.Email)
	if err != nil {
		log.Error().Msg("failed to find merchant")
		return nil, err
	}
	if existing != nil {
		log.Warn().Msg("merchant already exists")
		return nil, domain.ErrMerchantAlreadyExists
	}

	// create merchant
	merchant := mapRegisterMerchant(data)

	if err := s.merchantRepo.Create(ctx, merchant); err != nil {
		log.Error().Msg("failed to create merchant")
		return nil, err
	}

	// create api key
	apiKey, _, err := s.createAPIKey(ctx, merchant.ID)
	if err != nil {
		log.Error().Msg("failed to create api key")
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
	log := s.log.With().Str("func", "Activate").Str("email", email).Logger()

	existing, err := s.merchantRepo.FindByEmail(ctx, email)
	if err != nil {
		log.Error().Msg("failed to find merchant")
		return nil, err
	}
	if existing == nil {
		log.Warn().Msg("merchant not found")
		return nil, domain.ErrMerchantNotFound
	}
	if existing.Status != domain.MerchantStatusPending {
		log.Warn().Msg("invalid merchant status")
		return nil, domain.ErrMerchantStatusNotAccepted
	}

	values := map[string]interface{}{
		"status": domain.MerchantStatusActive,
	}

	merchant, err := s.merchantRepo.UpdateAndReturn(ctx, existing.ID, values)
	if err != nil {
		log.Error().Msg("failed to update merchant")
		return nil, err
	}

	return merchant, nil
}
