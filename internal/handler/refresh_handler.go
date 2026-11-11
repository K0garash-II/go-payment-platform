package handler

import (
	"context"
	"net/http"
	"time"

	"go-payment-platform/internal/auth"
)

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("failed to generate refresh token", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	newRefreshTokenHash := auth.HashRefreshToken(newRefreshToken)
	oldRefreshTokenHash := auth.HashRefreshToken(cookie.Value)

	now := time.Now()

	refreshTokenExpiresAt := now.Add(7 * 24 * time.Hour)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID, err := h.service.RotateRefreshToken(ctx, oldRefreshTokenHash, newRefreshTokenHash, refreshTokenExpiresAt)
	if err != nil {
		h.logger.Error("failed to rotate refresh token", "error", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.GenerateAccessToken(userID, h.jwtSecret)
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
		Secure:   false, //than change to https
		SameSite: http.SameSiteLaxMode,
		Expires:  now.Add(15 * time.Minute),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, //than change to https
		SameSite: http.SameSiteLaxMode,
		Expires:  refreshTokenExpiresAt,
	})

	w.WriteHeader(http.StatusOK)
}
