package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/validation"
)

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateAccountRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	owner, err := validation.ValidateOwner(input.Owner)
	if err != nil {
		http.Error(w, "invalid account owner", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	acc, err := h.service.CreateAccount(ctx, owner)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusCreated, acc); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		http.Error(w, "invalid account ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	acc, err := h.service.GetAccount(ctx, idInt)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, acc); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
