package handler

import (
	"errors"
	"net/http"

	"go-payment-platform/internal/payment"
)

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payment.ErrInvalidAmount):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, payment.ErrAccountNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)

	case errors.Is(err, payment.ErrInsufficientFunds):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, payment.ErrSameAccount):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, payment.ErrBalanceOverflow):
		http.Error(w, err.Error(), http.StatusBadRequest)

	case errors.Is(err, payment.ErrInvalidAccountID):
		http.Error(w, err.Error(), http.StatusBadRequest)

	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
