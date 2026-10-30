package handler

import (
	"encoding/json"
	"go-payment-platform/internal/dto"
	"net/http"
)

func (h *Handler) DepositAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	acc, err := h.service.Deposit(input.AccountID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(acc)

}
