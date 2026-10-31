package handler

import (
	"net/http"

	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/validation"
)

func (h *Handler) DepositAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.DepositRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	acc, err := h.service.Deposit(input.AccountID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, acc); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
