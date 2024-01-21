package handler

import (
	"encoding/json"
	"net/http"

	"github.com/la0rg/lingo/logger"
)

func JSON(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.WithContext(r.Context()).Error("unable to encode response", "error", err)
	}
}
