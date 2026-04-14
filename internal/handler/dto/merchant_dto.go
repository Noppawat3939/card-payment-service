package dto

type RegisterMerchantRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	WebhookURL *string `json:"webhook_url"`
}

type RegisterMerchantResponse struct {
	MerchantID string `json:"merchant_id"`
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	Status     string `json:"status"`
}

type UpdateMerchantStatusRequest struct {
	Email string `json:"email" binding:"required"`
}

type UpdateMerchantStatusResponse struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}
