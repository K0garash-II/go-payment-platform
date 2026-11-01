package payment

import (
	"context"
	"errors"

	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"
	"go-payment-platform/internal/transaction"
	"go-payment-platform/internal/user"
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
	RegisterUser(ctx context.Context, email string, passwordHash string) (user.User, account.Account, error)
	GetAccount(ctx context.Context, id int) (account.Account, error)
	Deposit(ctx context.Context, id int, amount int64) (account.Account, error)
	Transfer(ctx context.Context, fromID, toID int, amount int64) error
	GetTransactions(ctx context.Context, accountID, offset, limit int) ([]transaction.Transaction, error)
}

func (s *Service) RegisterUser(ctx context.Context, email string, passwordHash string) (user.User, account.Account, error) {
	u, acc, err := s.store.RegisterUser(ctx, email, passwordHash)

	if errors.Is(err, storage.ErrEmailAlreadyExists) {
		return user.User{}, account.Account{}, ErrEmailAlreadyExists
	}
	if err != nil {
		return user.User{}, account.Account{}, err
	}

	return u, acc, nil
}

func (s *Service) GetAccount(ctx context.Context, id int) (account.Account, error) {
	if id <= 0 {
		return account.Account{}, ErrInvalidAccountID
	}

	acc, err := s.store.GetAccount(ctx, id)

	if errors.Is(err, storage.ErrAccountNotFound) {
		return account.Account{}, ErrAccountNotFound
	}
	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Service) Deposit(ctx context.Context, id int, amount int64) (account.Account, error) {
	if amount <= 0 {
		return account.Account{}, ErrInvalidAmount
	}

	if id <= 0 {
		return account.Account{}, ErrInvalidAccountID
	}

	acc, err := s.store.Deposit(ctx, id, amount)
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

func (s *Service) Transfer(ctx context.Context, fromID, toID int, amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if fromID <= 0 || toID <= 0 {
		return ErrInvalidAccountID
	}

	if fromID == toID {
		return ErrSameAccount
	}

	err := s.store.Transfer(ctx, fromID, toID, amount)
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

func (s *Service) GetTransactions(ctx context.Context, accountID, offset, limit int) ([]transaction.Transaction, error) {
	if accountID <= 0 {
		return nil, ErrInvalidAccountID
	}

	_, err := s.store.GetAccount(ctx, accountID)
	if errors.Is(err, storage.ErrAccountNotFound) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	accTrans, err := s.store.GetTransactions(ctx, accountID, offset, limit)
	if err != nil {
		return nil, err
	}

	return accTrans, nil
}
