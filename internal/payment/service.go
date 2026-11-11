package payment

import (
	"context"
	"errors"
	"time"

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
	GetUserByEmail(ctx context.Context, email string) (user.User, error)
	GetAccount(ctx context.Context, id int, userID int) (account.Account, error)
	Deposit(ctx context.Context, id int, amount int64, userID int) (account.Account, error)
	Transfer(ctx context.Context, fromID, toID int, amount int64, userID int) error
	GetTransactions(ctx context.Context, accountID, offset, limit int, userID int) ([]transaction.Transaction, error)
	CreateSession(ctx context.Context, userID int, tokenHash string, sessionExpiresAt time.Time, tokenExpiresAt time.Time) error
	RotateRefreshToken(ctx context.Context, tokenHash string, newTokenHash string, tokenExpiresAt time.Time) (int, error)
}

func (s *Service) RegisterUser(
	ctx context.Context,
	email string, passwordHash string,
) (user.User, account.Account, error) {

	u, acc, err := s.store.RegisterUser(ctx, email, passwordHash)

	if errors.Is(err, storage.ErrEmailAlreadyExists) {
		return user.User{}, account.Account{}, ErrEmailAlreadyExists
	}
	if err != nil {
		return user.User{}, account.Account{}, err
	}

	return u, acc, nil
}

func (s *Service) GetUserByEmail(
	ctx context.Context, email string,
) (user.User, error) {

	u, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, storage.ErrUserNotFound) {
		return user.User{}, ErrUserNotFound
	}

	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

func (s *Service) GetAccount(
	ctx context.Context,
	id int, userID int,
) (account.Account, error) {

	if id <= 0 {
		return account.Account{}, ErrInvalidAccountID
	}

	if userID <= 0 {
		return account.Account{}, ErrInvalidUserID
	}

	acc, err := s.store.GetAccount(ctx, id, userID)

	if errors.Is(err, storage.ErrAccountNotFound) {
		return account.Account{}, ErrAccountNotFound
	}
	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Service) Deposit(
	ctx context.Context, id int,
	amount int64, userID int,
) (account.Account, error) {

	if amount <= 0 {
		return account.Account{}, ErrInvalidAmount
	}

	if id <= 0 {
		return account.Account{}, ErrInvalidAccountID
	}

	if userID <= 0 {
		return account.Account{}, ErrInvalidUserID
	}

	acc, err := s.store.Deposit(ctx, id, amount, userID)
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

func (s *Service) Transfer(
	ctx context.Context, fromID, toID int,
	amount int64, userID int,
) error {

	if amount <= 0 {
		return ErrInvalidAmount
	}

	if fromID <= 0 || toID <= 0 {
		return ErrInvalidAccountID
	}

	if fromID == toID {
		return ErrSameAccount
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	err := s.store.Transfer(ctx, fromID, toID, amount, userID)
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

func (s *Service) GetTransactions(
	ctx context.Context, accountID,
	offset, limit int, userID int,
) ([]transaction.Transaction, error) {

	if accountID <= 0 {
		return nil, ErrInvalidAccountID
	}

	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	_, err := s.store.GetAccount(ctx, accountID, userID)
	if errors.Is(err, storage.ErrAccountNotFound) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	accTrans, err := s.store.GetTransactions(ctx, accountID, offset, limit, userID)
	if err != nil {
		return nil, err
	}

	return accTrans, nil
}

func (s *Service) CreateSession(
	ctx context.Context, userID int, tokenHash string,
	sessionExpiresAt time.Time, tokenExpiresAt time.Time,
) error {
	if userID <= 0 {
		return ErrInvalidUserID
	}

	if tokenHash == "" {
		return ErrInvalidSession
	}

	now := time.Now()

	if !sessionExpiresAt.After(now) {
		return ErrInvalidSession
	}

	if !tokenExpiresAt.After(now) {
		return ErrInvalidSession
	}

	if tokenExpiresAt.After(sessionExpiresAt) {
		return ErrInvalidSession
	}

	return s.store.CreateSession(ctx, userID, tokenHash, sessionExpiresAt, tokenExpiresAt)
}

func (s *Service) RotateRefreshToken(
	ctx context.Context,
	tokenHash, newTokenHash string,
	tokenExpiresAt time.Time,
) (int, error) {

	if tokenHash == "" || newTokenHash == "" {
		return 0, ErrInvalidSession
	}

	now := time.Now()

	if !tokenExpiresAt.After(now) {
		return 0, ErrInvalidSession
	}

	if tokenHash == newTokenHash {
		return 0, ErrInvalidSession
	}

	userID, err := s.store.RotateRefreshToken(ctx, tokenHash, newTokenHash, tokenExpiresAt)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
