package dto

type TransferRequest struct {
	FromID int   `json:"from_id"`
	ToID   int   `json:"to_id"`
	Amount int64 `json:"amount"`
}
