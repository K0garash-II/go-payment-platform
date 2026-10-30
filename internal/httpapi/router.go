package httpapi

import (
	"log/slog"
	"net/http"

	"go-payment-platform/internal/handler"
	"go-payment-platform/internal/middleware"
	"go-payment-platform/internal/payment"
)

func NewRouter(p *payment.Service, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	h := handler.NewHandler(p, logger)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Error("failed to write health response", "error", err)
		}
	})

	mux.HandleFunc("POST /accounts", h.CreateAccount)
	mux.HandleFunc("GET /accounts/{id}", h.GetAccount)
	mux.HandleFunc("POST /deposits", h.DepositAccount)
	mux.HandleFunc("POST /transfers", h.TransferAccount)
	mux.HandleFunc("GET /accounts/{id}/transactions", h.GetAccountTransactions)

	return middleware.LoggingMiddleware(logger)(mux)
}
