package dto

type DepositRequest struct {
	AccountID int   `json:"account_id"`
	Amount    int64 `json:"amount"`
}
