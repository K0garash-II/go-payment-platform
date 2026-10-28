package account

type Account struct {
	ID      int     `json:"id"`
	Owner   string  `json:"owner"`
	Balance float64 `json:"balance"`
}