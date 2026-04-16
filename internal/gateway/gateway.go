package gateway

import "context"

type AuthorizeRequest struct {
	Amount   int64
	Currency string
	OrderID  string
}

type AuthorizeResponse struct {
	GatewayRef string
	Status     string
}

type TokenizeRequest struct {
	CardNumber  string
	ExpiryMonth string
	ExpiryYear  string
	CVV         string
}

type TokenizeResponse struct {
	CardToken string
	LastFour  string
	Brand     string
}

type CaptureRequest struct {
	GatewayRef string
	OrderID    string
}

type CaptureResponse struct {
	Status     string
	GatewayRef string
}

type Gateway interface {
	Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error)
	Capture(ctx context.Context, req CaptureRequest) (*CaptureResponse, error)
	Refund(ctx context.Context, gatewayRef string) error
	TokenizeCard(ctx context.Context, req TokenizeRequest) (*TokenizeResponse, error)
}
