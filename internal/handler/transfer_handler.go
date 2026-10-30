package handler

import (
	"go-payment-platform/internal/dto"
	"net/http"
)

func (h *Handler) TransferAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.TransferRequest
	
	if !decodeRequest(w, r, &input) {
		return
	}

	err := h.service.Transfer(input.FromID, input.ToID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
