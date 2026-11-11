package storage

import "errors"

var (
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrAccountNotFound    = errors.New("account not found")
	ErrBalanceOverflow    = errors.New("balance overflow")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidSession     = errors.New("invalid session")
)
