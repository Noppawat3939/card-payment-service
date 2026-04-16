package dto

type AuthorizePaymentRequest struct {
	CardNumber  string  `json:"card_number" binding:"required"`
	ExpiryMonth string  `json:"expiry_month" binding:"required"`
	ExpiryYear  string  `json:"expiry_year" binding:"required"`
	CVV         string  `json:"cvv" binding:"required"`
	Amount      int64   `json:"amount" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	Description *string `json:"description"`
}

type AuthorizePaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}

type CapturePaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	CapturedAt    string `json:"captured_at"`
}
