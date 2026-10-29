package httpapi

import (
	"encoding/json"
	"errors"
	"go-payment-platform/internal/payment"
	"net/http"
	"strconv"
	"strings"
)

type CreateAccountRequest struct {
	Owner string `json:"owner"`
}

type DepositRequest struct {
    AccountID int   `json:"account_id"`
    Amount    int64 `json:"amount"`
}

type TransferRequest struct {
	FromID int   `json:"from_id"`
	ToID   int   `json:"to_id"`
	Amount int64 `json:"amount"`
}

func check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func CreateAccount(p *payment.Service) http.HandlerFunc {
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

		acc := p.CreateAccount(input.Owner)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(acc)
	}
}

func GetAccount(p *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "invalid account ID", http.StatusBadRequest)
			return
		}

		acc, err := p.GetAccount(idInt)
		if err != nil {
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
		var input DepositRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		acc, err := p.Deposit(input.AccountID, input.Amount)

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

func TransferAccount(p *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input TransferRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		err := p.Transfer(input.FromID, input.ToID, input.Amount)
		if errors.Is(err, payment.ErrInvalidAmount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, payment.ErrAccountNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, payment.ErrInsufficientFunds) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, payment.ErrSameAccount) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetAccountTransactions(p *payment.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "invalid account ID", http.StatusBadRequest)
			return
		}

		accTrans, err := p.GetTransactions(idInt)
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
		json.NewEncoder(w).Encode(accTrans)

	}
}

func NewRouter(p *payment.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", check)
	mux.HandleFunc("POST /accounts", CreateAccount(p))
	mux.HandleFunc("GET /accounts/{id}", GetAccount(p))
	mux.HandleFunc("POST /deposits", DepositAccount(p))
	mux.HandleFunc("GET /accounts/{id}/transactions", GetAccountTransactions(p))
	mux.HandleFunc("POST /transfers", TransferAccount(p))

	return mux
}
