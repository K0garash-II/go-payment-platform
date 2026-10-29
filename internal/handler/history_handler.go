package handler

import (
	"encoding/json"
	"errors"
	"go-payment-platform/internal/payment"
	"net/http"
	"strconv"
)

func (h *Handler) GetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid account ID", http.StatusBadRequest)
		return
	}

	accTrans, err := h.service.GetTransactions(idInt)
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
	json.NewEncoder(w).Encode(accTrans)
}
