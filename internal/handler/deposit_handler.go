package handler

import (
	"context"
	"net/http"
	"time"

	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/middleware"
	"go-payment-platform/internal/validation"
)

func (h *Handler) DepositAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.DepositRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	acc, err := h.service.Deposit(ctx, input.AccountID, input.Amount, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, acc); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
