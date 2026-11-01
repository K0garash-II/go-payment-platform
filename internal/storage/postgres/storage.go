package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"

	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"
	"go-payment-platform/internal/transaction"
	"go-payment-platform/internal/user"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func (s *Store) RegisterUser(ctx context.Context, email string, passwordHash string) (user.User, account.Account, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return user.User{}, account.Account{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var pgErr *pgconn.PgError

	var u user.User

	err = tx.QueryRow(
		ctx,
		`INSERT INTO users (email, password_hash)
         VALUES ($1, $2)
         RETURNING id, email, password_hash, created_at`,
		email,
		passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return user.User{}, account.Account{}, storage.ErrEmailAlreadyExists
	}

	if err != nil {
		return user.User{}, account.Account{}, err
	}

	var acc account.Account

	err = tx.QueryRow(
		ctx,
		"INSERT INTO accounts (user_id) VALUES ($1) RETURNING id, user_id, balance",
		u.ID,
	).Scan(&acc.ID, &acc.UserID, &acc.Balance)

	if err != nil {
		return user.User{}, account.Account{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return user.User{}, account.Account{}, err
	}

	return u, acc, nil
}

func (s *Store) GetAccount(ctx context.Context, id int) (account.Account, error) {
	return scanAccount(s.pool.QueryRow(
		ctx,
		"SELECT id, user_id, balance FROM accounts WHERE id = $1",
		id,
	))
}

func (s *Store) Deposit(ctx context.Context, id int, amount int64) (account.Account, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return account.Account{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	acc, err := scanAccount(tx.QueryRow(
		ctx,
		`UPDATE accounts
		 SET balance = balance + $1
		 WHERE id = $2
		 AND balance <= 9223372036854775807 - $1
		 RETURNING id, user_id, balance`,
		amount, id,
	))

	if errors.Is(err, storage.ErrAccountNotFound) {
		var exists bool

		err = tx.QueryRow(
			ctx,
			"SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)",
			id,
		).Scan(&exists)

		if err != nil {
			return account.Account{}, err
		}

		if !exists {
			return account.Account{}, storage.ErrAccountNotFound
		}

		return account.Account{}, storage.ErrBalanceOverflow
	}

	if err != nil {
		return account.Account{}, err
	}

	if err := insertTransaction(ctx, tx, acc.ID, "deposit", amount); err != nil {
		return account.Account{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Store) Transfer(ctx context.Context, fromID, toID int, amount int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var accFrom, accTo account.Account

	firstID, secondID := fromID, toID

	if fromID > toID {
		firstID, secondID = toID, fromID
	}

	first, err := scanAccount(tx.QueryRow(
		ctx,
		"SELECT id, user_id, balance FROM accounts WHERE id = $1 FOR UPDATE",
		firstID,
	))
	if err != nil {
		return err
	}

	second, err := scanAccount(tx.QueryRow(
		ctx,
		"SELECT id, user_id, balance FROM accounts WHERE id = $1 FOR UPDATE",
		secondID,
	))
	if err != nil {
		return err
	}

	if first.ID == fromID {
		accFrom, accTo = first, second
	} else {
		accFrom, accTo = second, first
	}

	if accFrom.Balance < amount {
		return storage.ErrInsufficientFunds
	}

	_, err = tx.Exec(
		ctx,
		"UPDATE accounts SET balance = balance - $1 WHERE id = $2",
		amount, accFrom.ID,
	)
	if err != nil {
		return err
	}

	tag, err := tx.Exec(
		ctx,
		`UPDATE accounts SET balance = balance + $1 WHERE id = $2 AND balance <= $3::bigint - $1::bigint`,
		amount, accTo.ID, int64(math.MaxInt64),
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrBalanceOverflow
	}

	if err := insertTransaction(ctx, tx, accFrom.ID, "transfer_out", amount); err != nil {
		return err
	}

	if err := insertTransaction(ctx, tx, accTo.ID, "transfer_in", amount); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetTransactions(ctx context.Context, accountID, offset, limit int) ([]transaction.Transaction, error) {
	var history []transaction.Transaction

	rows, err := s.pool.Query(
		ctx,
		`SELECT id, account_id, type, amount, created_at
     FROM transactions
     WHERE account_id = $1
     ORDER BY created_at DESC, id DESC
     LIMIT $2 OFFSET $3`,
		accountID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var trans transaction.Transaction

		err := rows.Scan(
			&trans.ID,
			&trans.AccountID,
			&trans.Type,
			&trans.Amount,
			&trans.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		history = append(history, trans)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func NewStore(ctx context.Context, connString string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}
