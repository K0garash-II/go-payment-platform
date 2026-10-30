package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-payment-platform/internal/dto"
)

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateAccountRequest

	if err := decodeJSON(w, r, &input); err != nil {
		if errors.Is(err, ErrInvalidJSON) {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrBodyTooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	owner, err := validateOwner(input.Owner)
	if err != nil {
		if errors.Is(err, ErrInvalidOwner) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	acc := h.service.CreateAccount(owner)

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
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(acc)
}
