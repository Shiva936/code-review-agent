package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

type promptVersionResponse struct {
	ID         int    `json:"id"`
	Iteration  int    `json:"iteration"`
	PromptText string `json:"prompt_text"`
	RulesJSON  string `json:"rules_json"`
	Source     string `json:"source"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

type promptDeltaResponse struct {
	ID               int    `json:"id"`
	Iteration        int    `json:"iteration"`
	WeakestIssue     string `json:"weakest_issue"`
	InputJSON        string `json:"input_json"`
	RawOutput        string `json:"raw_output"`
	DeltaJSON        string `json:"delta_json"`
	ValidationStatus string `json:"validation_status"`
	Applied          bool   `json:"applied"`
	Source           string `json:"source"`
	Reason           string `json:"reason"`
	CreatedAt        string `json:"created_at"`
}

type promptArtifactsResponse struct {
	RunGroupID int                     `json:"run_group_id"`
	Versions   []promptVersionResponse `json:"versions"`
	Deltas     []promptDeltaResponse   `json:"deltas"`
}

func runGroupPromptArtifactsHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	_ = cfg
	groupID, err := strconv.Atoi(r.URL.Query().Get("group_id"))
	if err != nil || groupID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "valid group_id is required"})
		return
	}

	versions, err := storage.GetRunGroupPromptVersions(groupID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load prompt versions: %v", err)})
		return
	}
	deltas, err := storage.GetRunGroupPromptDeltas(groupID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load prompt deltas: %v", err)})
		return
	}

	resp := promptArtifactsResponse{RunGroupID: groupID}
	resp.Versions = make([]promptVersionResponse, 0, len(versions))
	for _, v := range versions {
		resp.Versions = append(resp.Versions, promptVersionResponse{
			ID:         v.ID,
			Iteration:  v.Iteration,
			PromptText: v.PromptText,
			RulesJSON:  v.RulesJSON,
			Source:     v.Source,
			Reason:     v.Reason,
			CreatedAt:  v.CreatedAt,
		})
	}
	resp.Deltas = make([]promptDeltaResponse, 0, len(deltas))
	for _, d := range deltas {
		resp.Deltas = append(resp.Deltas, promptDeltaResponse{
			ID:               d.ID,
			Iteration:        d.Iteration,
			WeakestIssue:     d.WeakestIssue,
			InputJSON:        d.InputJSON,
			RawOutput:        d.RawOutput,
			DeltaJSON:        d.DeltaJSON,
			ValidationStatus: d.ValidationStatus,
			Applied:          d.Applied,
			Source:           d.Source,
			Reason:           d.Reason,
			CreatedAt:        d.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
