package account

type Account struct {
	ID      int     `json:"id"`
	Owner   string  `json:"owner"`
	Balance int64 `json:"balance"`
}