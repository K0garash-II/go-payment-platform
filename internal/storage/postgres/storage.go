package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"
	"go-payment-platform/internal/transaction"
	"math"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func (s *Store) CreateAccount(ctx context.Context, owner string) (account.Account, error) {
	var acc account.Account
	err := s.pool.QueryRow(
		ctx,
		"INSERT INTO accounts (owner) VALUES ($1) RETURNING id, owner, balance",
		owner,
	).Scan(&acc.ID, &acc.Owner, &acc.Balance)

	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Store) GetAccount(ctx context.Context, id int) (account.Account, error) {
	var acc account.Account
	err := s.pool.QueryRow(
		ctx,
		"SELECT id, owner, balance FROM accounts WHERE id = $1",
		id,
	).Scan(&acc.ID, &acc.Owner, &acc.Balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return account.Account{}, storage.ErrAccountNotFound
	}

	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func (s *Store) Deposit(ctx context.Context, id int, amount int64) (account.Account, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return account.Account{}, err
	}
	defer tx.Rollback(ctx)

	var acc account.Account
	err = tx.QueryRow(
		ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2 RETURNING id, owner, balance",
		amount, id,
	).Scan(&acc.ID, &acc.Owner, &acc.Balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return account.Account{}, storage.ErrAccountNotFound
	}

	if err != nil {
		return account.Account{}, err
	}

	_, err = tx.Exec(
		ctx,
		"INSERT INTO transactions (account_id, type, amount) VALUES ($1, $2, $3)",
		acc.ID, "deposit", amount,
	)

	if err != nil {
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
	defer tx.Rollback(ctx)

	var accFrom, accTo account.Account

	firstID, secondID := fromID, toID

	if fromID > toID {
		firstID, secondID = toID, fromID
	}

	var first, second account.Account

	err = tx.QueryRow(
		ctx,
		"SELECT id, owner, balance FROM accounts WHERE id = $1 FOR UPDATE",
		firstID,
	).Scan(&first.ID, &first.Owner, &first.Balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	err = tx.QueryRow(
		ctx,
		"SELECT id, owner, balance FROM accounts WHERE id = $1 FOR UPDATE",
		secondID,
	).Scan(&second.ID, &second.Owner, &second.Balance)

	if errors.Is(err, pgx.ErrNoRows) {
		return storage.ErrAccountNotFound
	}
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

	_, err = tx.Exec(
		ctx,
		"INSERT INTO transactions (account_id, type, amount) VALUES ($1, $2, $3)",
		accFrom.ID, "transfer_out", amount,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		"INSERT INTO transactions (account_id, type, amount) VALUES ($1, $2, $3)",
		accTo.ID, "transfer_in", amount,
	)
	if err != nil {
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
