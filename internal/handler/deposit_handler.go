package handler

import (
	"encoding/json"
	"errors"
	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/payment"
	"net/http"
)

func (h *Handler) DepositAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	acc, err := h.service.Deposit(input.AccountID, input.Amount)

	if errors.Is(err, payment.ErrInvalidAmount) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, payment.ErrAccountNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(acc)

}
