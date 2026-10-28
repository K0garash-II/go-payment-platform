package payment

import (
	"errors"
	"go-payment-platform/internal/account"
)
var ErrInvalidAmount = errors.New("amount must be positive")
var ErrAccountNotFound = errors.New("account not found")

type Service struct {
    store AccountStore
}

func NewService(store AccountStore) *Service {
    return &Service{
        store: store,
    }
}

type AccountStore interface {
	Deposit(id int, amount int64) (account.Account, bool)
}

func (s *Service) Deposit(id int, amount int64) (account.Account, error) {
	if amount <= 0 {
		return account.Account{}, ErrInvalidAmount
	}

	acc, found := s.store.Deposit(id, amount)
	if !found {
		return account.Account{}, ErrAccountNotFound
	}

	return acc, nil
}

