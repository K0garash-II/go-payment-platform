package main

import (
	"log/slog"
	"os"

	"go-payment-platform/internal/httpapi"
	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/memory"

	"net/http"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store := memory.NewStore()
	p := payment.NewService(store)
	router := httpapi.NewRouter(p, logger)

	logger.Info("Starting server", "address", ":8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
