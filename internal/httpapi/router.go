package httpapi

import (
	"encoding/json"
	"errors"
	"go-payment-platform/internal/payment"
	"go-payment-platform/internal/storage/memory"
	"net/http"
	"strconv"
	"strings"
)

type CreateAccountRequest struct {
	Owner string `json:"owner"`
}

type DepositRequest struct {
	Amount int64 `json:"amount"`
}

func check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"service":"payment-api","status":"running"}`))
}

func CreateAccount(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input CreateAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		input.Owner = strings.TrimSpace(input.Owner)
		if input.Owner == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		acc := store.CreateAccount(input.Owner)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	}
}

func GetAccount(store *memory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "invalid account ID", http.StatusBadRequest)
			return
		}

		acc, found := store.GetAccount(idInt)
		if !found {
			http.Error(w, "Account not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(acc)
	}
}

func DepositAccount(p *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "invalid account ID", http.StatusBadRequest)
			return
		}

		var input DepositRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		acc, err := p.Deposit(idInt, input.Amount)

		if errors.Is(err, payment.ErrInvalidAmount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, payment.ErrAccountNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(acc)
	}
}

func NewRouter(store *memory.Store, p *payment.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", check)
	mux.HandleFunc("GET /", root)
	mux.HandleFunc("POST /accounts", CreateAccount(store))
	mux.HandleFunc("GET /accounts/{id}", GetAccount(store))
	mux.HandleFunc("POST /accounts/{id}/deposit", DepositAccount(p))

	return mux
}
