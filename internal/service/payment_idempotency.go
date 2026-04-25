package service

import (
	"card-payment-service/internal/domain"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func (s *PaymentService) cachedIdempotency(ctx context.Context, key, merchantID uuid.UUID, dest any, log zerolog.Logger) bool {
	idem, err := s.idemRepo.FindByKeyAndMerchantID(ctx, key, merchantID)

	if err != nil || idem == nil {
		return false
	}

	if err := json.Unmarshal(idem.Response, &dest); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal idempotency response")
		return false
	}

	log.Warn().Msg("duplicate idempotency key — returning cached response")
	return true
}

func (s *PaymentService) saveIdempotency(key, merchantID uuid.UUID, dest any, log zerolog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := json.Marshal(dest)
	if err != nil {
		log.Error().Err(err).Msg("failed to json marshall output")
		return
	}

	if err = s.idemRepo.Create(ctx, &domain.IdempotencyKey{
		Key:        key,
		MerchantID: merchantID,
		Response:   raw,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 1d
	}); err != nil {
		log.Warn().Err(err).Msg("failed to save idempotency key")
	}
}
