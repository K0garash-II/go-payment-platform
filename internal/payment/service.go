package payment

import (
	"errors"
	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"
	"go-payment-platform/internal/transaction"
)

type Service struct {
	store AccountStore
}

func NewService(store AccountStore) *Service {
	return &Service{
		store: store,
	}
}

type AccountStore interface {
	CreateAccount(owner string) account.Account
	GetAccount(id int) (account.Account, error)
	Deposit(id int, amount int64) (account.Account, error)
	Transfer(fromID, toID int, amount int64) error
	GetTransactions(accountID int) []transaction.Transaction
}

func (s *Service) CreateAccount(owner string) account.Account {
	acc := s.store.CreateAccount(owner)
	return acc
}

func (s *Service) GetAccount(id int) (account.Account, error) {
	acc, err := s.store.GetAccount(id)
	if errors.Is(err, storage.ErrAccountNotFound) {
		return account.Account{}, ErrAccountNotFound
	}
	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Service) Deposit(id int, amount int64) (account.Account, error) {
	if amount <= 0 {
		return account.Account{}, ErrInvalidAmount
	}

	acc, err := s.store.Deposit(id, amount)
	if errors.Is(err, storage.ErrAccountNotFound) {
		return account.Account{}, ErrAccountNotFound
	}
	if errors.Is(err, storage.ErrBalanceOverflow) {
		return account.Account{}, ErrBalanceOverflow
	}
	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Service) Transfer(fromID, toID int, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if fromID == toID {
		return ErrSameAccount
	}

	err := s.store.Transfer(fromID, toID, amount)
	if errors.Is(err, storage.ErrAccountNotFound) {
		return ErrAccountNotFound
	}
	if errors.Is(err, storage.ErrInsufficientFunds) {
		return ErrInsufficientFunds
	}
	if errors.Is(err, storage.ErrBalanceOverflow) {
		return ErrBalanceOverflow
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetTransactions(accountID int) ([]transaction.Transaction, error) {
	_, err := s.store.GetAccount(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	accTrans := s.store.GetTransactions(accountID)

	return accTrans, nil
}
