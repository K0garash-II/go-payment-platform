package handler

import (
	"encoding/json"
	"net/http"

	"go-payment-platform/internal/dto"
)

func (h *Handler) DepositAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.DepositRequest

	if !decodeRequest(w, r, &input) {
		return
	}

	acc, err := h.service.Deposit(input.AccountID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(acc)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
