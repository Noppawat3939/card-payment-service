package service

type RegisterInput struct {
	Name       string
	Email      string
	WebhookURL *string
}

type RegisterOutput struct {
	MerchantID string `json:"merchant_id"`
	APIKey     string `json:"api_key"`
	Status     string `json:"status"`
}
