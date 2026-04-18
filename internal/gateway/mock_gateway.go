package gateway

import (
	"card-payment-service/internal/domain"
	"context"
	"os"
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
		return nil, domain.ErrCardAmoutInvalid
	}
	if req.OrderID == "" || req.Currency == "" || len(req.Currency) != 3 {
		return nil, domain.ErrCardInforInvalid
	}
	// simulate declined card
	if req.Amount == 99999 {
		return nil, domain.ErrCardDeclinded
	}
	// simulate insufficient funds
	if req.Amount == 9999 {
		return nil, domain.ErrInsufficientFunds
	}

	return &AuthorizeResponse{
		GatewayRef: "gw_mock_001",
		Status:     "authorized",
	}, nil
}

func (m *MockGateway) Capture(ctx context.Context, req CaptureRequest) (*CaptureResponse, error) {
	if req.GatewayRef == "" {
		return nil, domain.ErrInvalidGatewayRef
	}
	// simulate rejected
	if strings.Contains(req.GatewayRef, "FAIL") {
		return nil, domain.ErrCardCaptureFailed
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

func (m *MockGateway) Void(ctx context.Context, req VoidRequest) (*VoidResponse, error) {
	if req.GatewayRef == "" {
		return nil, domain.ErrInvalidGatewayRef
	}

	// simulate rejected
	if strings.Contains(req.GatewayRef, "VOID_FAIL") {
		return nil, domain.ErrCardCaptureFailed
	}

	// simulate timeout
	if strings.Contains(req.GatewayRef, "VOID_TIMEOUT") {
		return nil, context.DeadlineExceeded
	}

	return &VoidResponse{
		Status:     "voided",
		GatewayRef: req.GatewayRef,
	}, nil
}

func (m *MockGateway) Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error) {
	if req.GatewayRef == "" {
		return nil, domain.ErrInvalidGatewayRef
	}

	if req.Amount <= 0 {
		return nil, domain.ErrCardAmoutInvalid
	}

	// simulate rejected
	if strings.Contains(req.GatewayRef, "REFUND_FAIL") {
		return nil, domain.ErrCardCaptureFailed
	}
	// simulate timeout
	if strings.Contains(req.GatewayRef, "REFUND_TIMEOUT") {
		return nil, context.DeadlineExceeded
	}

	return &RefundResponse{
		RefundRef: "rf_" + uuid.NewString(),
		Status:    "processing",
	}, nil
}

func (*MockGateway) TokenizeCard(ctx context.Context, req TokenizeRequest) (*TokenizeResponse, error) {
	card := creditcard.Card{
		Number: req.CardNumber,
		Cvv:    req.CVV,
		Month:  req.ExpiryMonth,
		Year:   req.ExpiryYear,
	}

	isDev := os.Getenv("APP_ENV") == "develop"

	if isDev {
		if scenario, ok := GetTestCardScenario(req.CardNumber); ok {
			switch scenario {
			case CardScenarioDeclined:
				return nil, domain.ErrCardDeclinded
			case CardScenarioInsufficientFunds:
				return nil, domain.ErrInsufficientFunds
			case CardScenarioExpired:
				return nil, domain.ErrExpiredCard
			case CardScenarioSuccess:
				return buildSuccessResponse(req.CardNumber)
			}
		}

	}

	// validation card number
	if err := card.Validate(false); err != nil {
		return nil, err
	}

	return buildSuccessResponse(card.Number)
}

func buildSuccessResponse(cardNumber string) (*TokenizeResponse, error) {
	card := creditcard.Card{Number: cardNumber}

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
