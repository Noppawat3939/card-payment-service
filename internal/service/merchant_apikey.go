package service

import (
	"card-payment-service/pkg"
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

func (s *MerchantService) createAPIKey(ctx context.Context, merchantID uuid.UUID) (*string, error) {
	secret := generateSecret()
	hashed := pkg.HashSHA256(secret)
	apiKey := mapRegisterAPIKey(merchantID, secret[:16], string(hashed))

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return &secret, nil
}

func generateSecret() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "sk_live_" + hex.EncodeToString(b)
}
