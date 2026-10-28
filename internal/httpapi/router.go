package httpapi

import (
	"encoding/json"
	"go-payment-platform/internal/storage/memory"
	"net/http"
	"strconv"
	"strings"
)

type CreateAccountRequest struct {
	Owner string `json:"owner"`
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

func NewRouter(store *memory.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", check)
	mux.HandleFunc("GET /", root)
	mux.HandleFunc("POST /accounts", CreateAccount(store))
	mux.HandleFunc("GET /accounts/{id}", GetAccount(store))

	return mux
}
