package handler

import (
	"log/slog"

	"go-payment-platform/internal/payment"
)

type Handler struct {
	service   *payment.Service
	logger    *slog.Logger
	jwtSecret []byte
}

func NewHandler(service *payment.Service, logger *slog.Logger, jwtSecret string) *Handler {
	return &Handler{
		service:   service,
		logger:    logger,
		jwtSecret: []byte(jwtSecret),
	}
}
