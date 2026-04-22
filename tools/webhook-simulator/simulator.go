package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Simulator struct {
	targetURL  string
	secret     string
	httpClient *http.Client
}

func NewSimulator(targetURL, secret string) *Simulator {
	return &Simulator{
		targetURL: targetURL,
		secret:    secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type SimulateResult struct {
	Event      string `json:"event"`
	GatewayRef string `json:"gateway_ref"`
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

func (s *Simulator) Send(event EventType, gatewayRef string) (*SimulateResult, error) {
	// build payload
	payload, err := buildPayload(event, gatewayRef)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// sign HMAC and make request
	signature := s.sign(b)

	req, err := s.buildHttpRequest(b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", signature)

	result := SimulateResult{Event: string(event), GatewayRef: gatewayRef}

	// make request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		result.Success = false
		result.Error = err.Error()

		return &result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Success = resp.StatusCode >= 200 && resp.StatusCode < 300

	if !result.Success {
		result.Error = fmt.Sprintf("payment service returned %d", resp.StatusCode)
	}

	return &result, nil
}

func (s *Simulator) sign(b []byte) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(b)
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Simulator) buildHttpRequest(body []byte) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, s.targetURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	return req, nil
}
