package dto

type AuthorizePaymentRequest struct {
	Amount      int64   `json:"amount" binding:"required"`
	CardNumber  string  `json:"card_number" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	CVV         string  `json:"cvv" binding:"required"`
	Description *string `json:"description"`
	ExpiryMonth string  `json:"expiry_month" binding:"required"`
	ExpiryYear  string  `json:"expiry_year" binding:"required"`
}

type AuthorizePaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type CapturePaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type VoidPaymentResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}
