package memory

import (
	"go-payment-platform/internal/account"
	"go-payment-platform/internal/payment"
	"sync"
	"time"
)

type Store struct {
	Accounts          map[int]account.Account
	NextID            int
	Transactions      []payment.Transaction
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

func (s *Store) GetAccount(id int) (account.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, ok := s.Accounts[id]
	return acc, ok
}

func (s *Store) Deposit(id int, amount int64) (account.Account, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	acc, ok := s.Accounts[id]
	if !ok {
		return account.Account{}, false
	}

	acc.Balance += amount
	s.Accounts[id] = acc

	transaction := payment.Transaction{
		ID: s.NextTransactionID,
		AccountID: id,
		Type: "deposit",
		Amount: amount,
		CreatedAt: time.Now(),
	}

	s.Transactions = append(s.Transactions, transaction)

	s.NextTransactionID += 1

	return acc, true
}

func (s *Store) GetTransactions(accountID int) []payment.Transaction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	accTransactions := make([]payment.Transaction, 0)

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
		Transactions:      make([]payment.Transaction, 0),
		NextTransactionID: 1,
	}
}
