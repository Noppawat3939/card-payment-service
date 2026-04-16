package gateway

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"

	creditcard "github.com/durango/go-credit-card"
	"github.com/google/uuid"
)

type MockGateway struct{}

func NewMockGateway() Gateway {
	return &MockGateway{}
}

func (m *MockGateway) Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.New("amount invalid")
	}
	// simulate declined card
	if req.Amount == 99999 {
		return nil, errors.New("card_declined")
	}
	// simulate insufficient funds
	if req.Amount == 9999 {
		return nil, errors.New("insufficient_funds")
	}

	if req.OrderID == "" {
		return nil, errors.New("order_id is required")
	}
	if req.Currency == "" || len(req.Currency) != 3 {
		return nil, errors.New("currency invalid")
	}

	return &AuthorizeResponse{
		GatewayRef: "gw_mock_001",
		Status:     "authorized",
	}, nil
}

func (m *MockGateway) Capture(ctx context.Context, req CaptureRequest) (*CaptureResponse, error) {
	if req.GatewayRef == "" {
		return nil, errors.New("missing gateway reference")
	}
	// simulate rejected
	if strings.Contains(req.GatewayRef, "FAIL") {
		return nil, errors.New("capture_failed")
	}
	// simulate timeout
	if strings.Contains(req.GatewayRef, "TIMEOUT") {
		return nil, context.DeadlineExceeded
	}

	return &CaptureResponse{
		Status:     "captured",
		GatewayRef: req.GatewayRef,
	}, nil
}

func (m *MockGateway) Refund(ctx context.Context, gatewayRef string) error {
	return nil
}

func (*MockGateway) TokenizeCard(ctx context.Context, req TokenizeRequest) (*TokenizeResponse, error) {
	card := creditcard.Card{Number: req.CardNumber, Cvv: req.CVV, Month: req.ExpiryMonth, Year: req.ExpiryYear}
	testNumbers := []string{"4242424242424242"} // TODO: move to whitelist later
	isDev := os.Getenv("APP_ENV") == "develop"
	allowed := slices.Contains(testNumbers, card.Number) && isDev

	// validation card number
	if err := card.Validate(allowed); err != nil {
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
