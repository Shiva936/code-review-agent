package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

type runResponse struct {
	Iteration int    `json:"iteration"`
	Score     int    `json:"score"`
	Weakness  string `json:"weakness"`
}

func runsHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	runs, err := storage.GetRuns()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load runs: %v", err)})
		return
	}

	resp := make([]runResponse, 0, len(runs))
	for _, run := range runs {
		resp = append(resp, runResponse{
			Iteration: run.Iteration,
			Score:     run.Score,
			Weakness:  run.Weakness,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string][]runResponse{"runs": resp})
}
