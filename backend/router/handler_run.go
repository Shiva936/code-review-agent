package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

func runHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code   string `json:"code"`
		Prompt string `json:"prompt"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}
	if req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "`code` is required"})
		return
	}

	const iterations = 5
	basePrompt := `You are a strict code reviewer... Provide categorized feedback (logic, performance, security, style) with clear severities (critical, minor, suggestion) and actionable fixes. Avoid vague advice.`

	w.Header().Set("Content-Type", "application/json")
	groupID, err := storage.CreateRunGroup(req.Code, basePrompt, iterations, "pending")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to create run group: %v", err)})
		return
	}
	if err := storage.InitializeRunGroupRuns(groupID, iterations); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to initialize run iterations: %v", err)})
		return
	}

	// Start async processing
	go processRunGroupAsync(cfg, groupID, req.Code, req.Prompt)

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"run_group_id": groupID,
		"status":       "started",
	})
}
