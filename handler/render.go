package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/la0rg/lingo/logger"
	"github.com/la0rg/lingo/web"
)

func JSON(w http.ResponseWriter, r *http.Request, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.WithContext(r.Context()).Error("unable to encode response", "error", err)
	}
}

func (h *handlers) HTML(w http.ResponseWriter, r *http.Request, name web.TemplateName, data any) error {
	w.Header().Set("Content-Type", "text/html")
	html, err := h.tmpl.Render(name, data)
	if err != nil {
		return fmt.Errorf("rendering '%s' template failed: %v", name, err)
	}

	if _, err = w.Write(html); err != nil {
		return fmt.Errorf("unable to write response: %v", err)
	}

	return nil
}
