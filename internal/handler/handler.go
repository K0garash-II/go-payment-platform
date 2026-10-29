package handler

import (
	"log/slog"

	"go-payment-platform/internal/payment"
)

type Handler struct {
	service *payment.Service
	logger  *slog.Logger
}

func NewHandler(service *payment.Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}
