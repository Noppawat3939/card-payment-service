package gateway

import (
	"context"

	creditcard "github.com/durango/go-credit-card"
	"github.com/google/uuid"
)

type MockGateway struct{}

func NewMockGateway() Gateway {
	return &MockGateway{}
}

func (m *MockGateway) Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error) {
	return &AuthorizeResponse{GatewayRef: "gw_mock_001", Status: "authorized"}, nil
}

func (m *MockGateway) Capture(ctx context.Context, gatewayRef string) error {
	return nil
}

func (m *MockGateway) Refund(ctx context.Context, gatewayRef string) error {
	return nil
}

func (*MockGateway) TokenizeCard(ctx context.Context, req TokenizeRequest) (*TokenizeResponse, error) {
	card := creditcard.Card{Number: req.CardNumber}

	// validation card number
	if err := card.Validate(); err != nil {
		return nil, err
	}
	if err := card.ValidateExpiration(); err != nil {
		return nil, err
	}
	if err := card.ValidateCVV(); err != nil {
		return nil, err
	}

	lastFour, err := card.LastFour()
	if err != nil {
		return nil, err
	}

	return &TokenizeResponse{
		CardToken: "tok_" + uuid.NewString(),
		LastFour:  lastFour,
		Brand:     "VISA",
	}, nil
}
