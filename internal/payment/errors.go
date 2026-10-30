package payment

import "errors"

var (
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrAccountNotFound   = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrSameAccount       = errors.New("same account")
	ErrBalanceOverflow   = errors.New("balance overflow")
)
