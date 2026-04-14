package domain

import "errors"

var (
	ErrBodyInvalid               = errors.New("invalid body request")
	ErrMerchantAlreadyExists     = errors.New("merchant email already exists")
	ErrMerchantNotFound          = errors.New("merchant email not found")
	ErrMerchantStatusNotAccepted = errors.New("merchant current status not accepted")
)
