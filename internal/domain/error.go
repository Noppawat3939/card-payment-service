package domain

import "errors"

var (
	ErrBodyInvalid           = errors.New("invalid body request")
	ErrMerchantAlreadyExists = errors.New("merchant email already exists")
)
