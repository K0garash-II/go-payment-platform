package httpapi

import (
	"log/slog"
	"net/http"

	"go-payment-platform/internal/handler"
	"go-payment-platform/internal/middleware"
	"go-payment-platform/internal/payment"
)

func NewRouter(p *payment.Service, logger *slog.Logger, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	h := handler.NewHandler(p, logger, jwtSecret)

	authMiddleware := middleware.AuthMiddleware([]byte(jwtSecret))

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Error("failed to write health response", "error", err)
		}
	})

	mux.HandleFunc("POST /register", h.RegisterUser)
	mux.HandleFunc("POST /login", h.LoginUser)
	mux.HandleFunc("POST /refresh", h.RefreshToken)
	mux.Handle("GET /accounts/{id}", authMiddleware(http.HandlerFunc(h.GetAccount)))
	mux.Handle("POST /deposits", authMiddleware(http.HandlerFunc(h.DepositAccount)))
	mux.Handle("POST /transfers", authMiddleware(http.HandlerFunc(h.TransferAccount)))
	mux.Handle("GET /accounts/{id}/transactions", authMiddleware(http.HandlerFunc(h.GetAccountTransactions)))

	return middleware.LoggingMiddleware(logger)(mux)
}
