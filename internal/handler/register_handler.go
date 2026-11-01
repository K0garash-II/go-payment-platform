package handler

import (
	"context"
	"net/http"
	"time"

	"go-payment-platform/internal/account"
	"go-payment-platform/internal/auth"
	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/user"
	"go-payment-platform/internal/validation"
)

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var input dto.RegisterRequest

	if !validation.DecodeRequest(w, r, &input) {
		return
	}

	email, err := validation.ValidateEmail(input.Email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	password, err := validation.ValidatePassword(input.Password)
	if err != nil {
		http.Error(w, "invalid password", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		h.logger.Error("failed to hash password", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	u, acc, err := h.service.RegisterUser(ctx, email, passwordHash)
	if err != nil {
		handleError(w, err)
		return
	}

	response := struct {
		User    user.User       `json:"user"`
		Account account.Account `json:"account"`
	}{
		User:    u,
		Account: acc,
	}

	if err := WriteJSON(w, http.StatusCreated, response); err != nil {
		h.logger.Error("failed to write registration response", "error", err)
	}
}
