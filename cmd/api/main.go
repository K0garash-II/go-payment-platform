package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go-payment-platform/internal/config"
	"go-payment-platform/internal/httpapi"
	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/postgres"

	"net/http"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Config error", "error", err)
		os.Exit(1)
	}

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	store, err := postgres.NewStore(context.Background(), connString)
	if err != nil {
		logger.Error("DB error", "error", err)
		os.Exit(1)
	}

	defer store.Close()
	
	p := payment.NewService(store)
	router := httpapi.NewRouter(p, logger)

	logger.Info("Starting server", "address", ":8080")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}
