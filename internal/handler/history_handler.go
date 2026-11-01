package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"go-payment-platform/internal/validation"
)

func (h *Handler) GetAccountTransactions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		http.Error(w, "invalid account ID", http.StatusBadRequest)
		return
	}

	limit, offset, err := validation.ParsePagination(r)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrInvalidLimit):
			http.Error(w, "invalid limit value", http.StatusBadRequest)
		case errors.Is(err, validation.ErrInvalidPage):
			http.Error(w, "invalid page value", http.StatusBadRequest)
		case errors.Is(err, validation.ErrPageTooLarge):
			http.Error(w, "page too large", http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	accTrans, err := h.service.GetTransactions(ctx, idInt, offset, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	if err := WriteJSON(w, http.StatusOK, accTrans); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}
