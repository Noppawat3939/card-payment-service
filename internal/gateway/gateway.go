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

type VoidRequest struct {
	GatewayRef string
}

type VoidResponse struct {
	Status     string
	GatewayRef string
}

type RefundRequest struct {
	GatewayRef string
	Amount     int64
}

type RefundResponse struct {
	RefundRef string
	Status    string
}

type Gateway interface {
	Authorize(ctx context.Context, req AuthorizeRequest) (*AuthorizeResponse, error)
	Capture(ctx context.Context, req CaptureRequest) (*CaptureResponse, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResponse, error)
	Void(ctx context.Context, req VoidRequest) (*VoidResponse, error)
	TokenizeCard(ctx context.Context, req TokenizeRequest) (*TokenizeResponse, error)
}
