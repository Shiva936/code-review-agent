package router

import (
	"encoding/json"
	"net/http"

	"github.com/Shiva936/code-review-agent/backend/config"
)

func healthHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
