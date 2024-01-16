package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *handlers) GetCardHandler(w http.ResponseWriter, r *http.Request) error {
	result, err := h.cards.Generate(
		r.Context(), "de", "en",
		"Kuchen",
		"käse",
		"verrückt",
	)
	if err != nil {
		return fmt.Errorf("unable to generate cards: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return fmt.Errorf("unable to encode response: %v", err)
	}

	return nil
}
