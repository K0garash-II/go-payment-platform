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

	if !decodeRequest(w, r, &input) {
		return
	}

	owner, err := validateOwner(input.Owner)
	if err != nil {
		if errors.Is(err, ErrInvalidOwner) {
			http.Error(w, "invalid account owner", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	acc := h.service.CreateAccount(owner)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(acc)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid account ID", http.StatusBadRequest)
		return
	}
	if idInt <= 0 {
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
	err = json.NewEncoder(w).Encode(acc)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
