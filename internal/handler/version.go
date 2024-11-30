package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/config"
)

func (h *handlers) Version(w http.ResponseWriter, r *http.Request) error {
	JSON(w, r, config.GetBuildInfo())
	return nil
}
