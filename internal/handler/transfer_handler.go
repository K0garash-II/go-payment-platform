package handler

import (
	"context"
	"net/http"
	"time"

	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/middleware"
	"go-payment-platform/internal/validation"
)

func (h *Handler) TransferAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.TransferRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.service.Transfer(ctx, input.FromID, input.ToID, input.Amount, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
