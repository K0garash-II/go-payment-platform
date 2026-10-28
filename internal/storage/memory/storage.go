package memory

import (
	"go-payment-platform/internal/account"
	"sync"
)

type Store struct {
	Accounts map[int]account.Account
	NextID   int
	mu       sync.RWMutex
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

func NewStore() *Store {
	return &Store{
		Accounts: make(map[int]account.Account),
		NextID:   1,
	}
}
