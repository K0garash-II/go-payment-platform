package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (h *Handler) GetAccountTransactions(w http.ResponseWriter, r *http.Request) {
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

	accTrans, err := h.service.GetTransactions(idInt)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accTrans)
}
