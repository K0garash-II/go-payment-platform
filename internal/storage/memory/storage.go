package memory

import (
	"go-payment-platform/internal/account"
	"go-payment-platform/internal/storage"
	"go-payment-platform/internal/transaction"
	"sync"
	"time"
)

type Store struct {
	Accounts          map[int]account.Account
	NextID            int
	Transactions      []transaction.Transaction
	NextTransactionID int
	mu                sync.RWMutex
}

func (s *Store) CreateAccount(owner string) account.Account {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.NextID
	acc := account.Account{
		ID:      id,
		Owner:   owner,
		Balance: 0,
	}

	s.Accounts[id] = acc
	s.NextID += 1

	return acc
}

func (s *Store) GetAccount(id int) (account.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, ok := s.Accounts[id]
	if !ok {
		return account.Account{}, storage.ErrAccountNotFound
	}
	return acc, nil
}

func (s *Store) Deposit(id int, amount int64) (account.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, ok := s.Accounts[id]
	if !ok {
		return account.Account{}, storage.ErrAccountNotFound
	}

	acc.Balance += amount
	s.Accounts[id] = acc

	transaction := transaction.Transaction{
		ID:        s.NextTransactionID,
		AccountID: id,
		Type:      "deposit",
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	s.Transactions = append(s.Transactions, transaction)

	s.NextTransactionID += 1

	return acc, nil
}

func (s *Store) Transfer(fromID, toID int, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	accountFrom, ok := s.Accounts[fromID]
	if !ok {
		return storage.ErrAccountNotFound
	}
	accountTo, ok := s.Accounts[toID]
	if !ok {
		return storage.ErrAccountNotFound
	}

	if accountFrom.Balance < amount {
		return storage.ErrInsufficientFunds
	}

	accountFrom.Balance -= amount
	accountTo.Balance += amount

	s.Accounts[fromID] = accountFrom
	s.Accounts[toID] = accountTo

	transactionFrom := transaction.Transaction{
		ID:        s.NextTransactionID,
		AccountID: fromID,
		Type:      "transfer",
		Amount:    -amount,
		CreatedAt: time.Now(),
	}
	s.NextTransactionID += 1
	s.Transactions = append(s.Transactions, transactionFrom)

	transactionTo := transaction.Transaction{
		ID:        s.NextTransactionID,
		AccountID: toID,
		Type:      "transfer",
		Amount:    amount,
		CreatedAt: time.Now(),
	}
	s.NextTransactionID += 1
	s.Transactions = append(s.Transactions, transactionTo)

	return nil
}

func (s *Store) GetTransactions(accountID int) []transaction.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accTransactions := make([]transaction.Transaction, 0)

	for i := range s.Transactions {
		if s.Transactions[i].AccountID == accountID {
			accTransactions = append(accTransactions, s.Transactions[i])
		}
	}

	return accTransactions
}

func NewStore() *Store {
	return &Store{
		Accounts:          make(map[int]account.Account),
		NextID:            1,
		Transactions:      make([]transaction.Transaction, 0),
		NextTransactionID: 1,
	}
}
