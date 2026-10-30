package handler

import (
	"encoding/json"
	"go-payment-platform/internal/dto"
	"net/http"
)

func (h *Handler) TransferAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.service.Transfer(input.FromID, input.ToID, input.Amount)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
