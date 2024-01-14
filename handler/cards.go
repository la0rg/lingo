package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func (h *handlers) GetCardHandler(w http.ResponseWriter, r *http.Request) {
	result, err := h.cards.Generate(
		r.Context(), "de", "en",
		"Kuchen",
		"käse",
		"verrückt",
	)
	if err != nil {
		slog.Error("failed to generate cards", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		slog.Error("unable to encode response", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("GetCardHandler finished")
}
