package handler

import (
	"net/http"

	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/validation"
)

func (h *Handler) TransferAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.TransferRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	err := h.service.Transfer(input.FromID, input.ToID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
