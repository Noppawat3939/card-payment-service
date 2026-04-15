package gateway

import "context"

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
