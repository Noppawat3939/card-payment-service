package merchant

import (
	"card-payment-service/internal/domain"
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *MerchantService) createAPIKey(ctx context.Context, merchantID uuid.UUID) (*domain.APIKey, string, error) {
	secret := generateSecret()

	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	apiKey := mapRegisterAPIKey(merchantID, secret[:16], string(hashed))

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, secret, nil
}

func generateSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "sk_live_" + hex.EncodeToString(b)
}
