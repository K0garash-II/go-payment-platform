package httpapi

import (
	"log/slog"
	"net/http"

	"go-payment-platform/internal/handler"
	"go-payment-platform/internal/payment"
)

func NewRouter(p *payment.Service, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	h := handler.NewHandler(p, logger)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /accounts", h.CreateAccount)
	mux.HandleFunc("GET /accounts/{id}", h.GetAccount)
	mux.HandleFunc("POST /deposits", h.DepositAccount)
	mux.HandleFunc("POST /transfers", h.TransferAccount)
	mux.HandleFunc("GET /accounts/{id}/transactions", h.GetAccountTransactions)

	return mux
}
