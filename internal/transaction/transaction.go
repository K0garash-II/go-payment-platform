package transaction

import "time"

type Transaction struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Type      string    `json:"type"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
