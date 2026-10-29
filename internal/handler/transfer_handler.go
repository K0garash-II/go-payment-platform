package handler

import (
	"encoding/json"
	"errors"
	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/payment"
	"net/http"
)

func (h *Handler) TransferAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.service.Transfer(input.FromID, input.ToID, input.Amount)
	if errors.Is(err, payment.ErrInvalidAmount) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, payment.ErrAccountNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if errors.Is(err, payment.ErrInsufficientFunds) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errors.Is(err, payment.ErrSameAccount) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
