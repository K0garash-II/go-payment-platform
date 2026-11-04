package handler

import (
	"context"
	"net/http"
	"time"

	"go-payment-platform/internal/auth"
	"go-payment-platform/internal/dto"
	"go-payment-platform/internal/validation"
)

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginRequest

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

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	u, err := h.service.GetUserByEmail(ctx, email)
	if err != nil {
		handleError(w, err)
		return
	}

	err = auth.CheckPassword(password, u.PasswordHash)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.GenerateAccessToken(u.ID, h.jwtSecret)
	if err != nil {
		h.logger.Error("failed to generate access token", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,  //than change to https
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})

	w.WriteHeader(http.StatusOK)
}
