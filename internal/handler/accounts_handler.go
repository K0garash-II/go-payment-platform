package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-payment-platform/internal/dto"
)

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	input.Owner = strings.TrimSpace(input.Owner)
	if input.Owner == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	acc := h.service.CreateAccount(input.Owner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(acc)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid account ID", http.StatusBadRequest)
		return
	}

	acc, err := h.service.GetAccount(idInt)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(acc)
}
