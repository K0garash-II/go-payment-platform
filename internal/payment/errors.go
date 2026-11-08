package payment

import "errors"

var (
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrAccountNotFound    = errors.New("account not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrSameAccount        = errors.New("same account")
	ErrBalanceOverflow    = errors.New("balance overflow")
	ErrInvalidAccountID   = errors.New("invalid account id")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidUserID      = errors.New("invalid user id")
	ErrInvalidSession     = errors.New("invalid session")
)
