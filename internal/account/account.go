package account

type Account struct {
	ID      int   `json:"id"`
	UserID  int   `json:"user_id"`
	Balance int64 `json:"balance"`
}
