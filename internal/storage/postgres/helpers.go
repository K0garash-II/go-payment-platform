package postgres

import (
	"context"
	"errors"

	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"

	"github.com/jackc/pgx/v5"
)

func scanAccount(row pgx.Row) (account.Account, error) {
	var acc account.Account

	err := row.Scan(
		&acc.ID,
		&acc.Owner,
		&acc.Balance,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return account.Account{}, storage.ErrAccountNotFound
	}

	if err != nil {
		return account.Account{}, err
	}

	return acc, nil
}

func insertTransaction(ctx context.Context,
	tx pgx.Tx,
	accountID int,
	transactionType string,
	amount int64,
) error {
	_, err := tx.Exec(
		ctx,
		`INSERT INTO transactions (account_id, type, amount)
         VALUES ($1, $2, $3)`,
		accountID,
		transactionType,
		amount,
	)

	return err
}
