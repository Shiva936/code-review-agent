package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/models"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

type runGroupRunResponse struct {
	Iteration     int    `json:"iteration"`
	Score         int    `json:"score"`
	Weakness      string `json:"weakness"`
	Status        string `json:"status"`
	Progress      int    `json:"progress_percent,omitempty"`
	Actionability *int   `json:"actionability,omitempty"`
	Specificity   *int   `json:"specificity,omitempty"`
	Severity      *int   `json:"severity,omitempty"`
	Structure     *int   `json:"structure,omitempty"`
	Samples       []models.SampleEvalMetrics `json:"samples,omitempty"`
}

type runGroupResponse struct {
	ID        int                   `json:"id"`
	InputCode string                `json:"input_code"`
	Status    string                `json:"status"`
	CreatedAt string                `json:"created_at"`
	UpdatedAt string                `json:"updated_at"`
	Iterations []runGroupRunResponse `json:"iterations"`
}

type runGroupsResponse struct {
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Groups   []runGroupResponse `json:"groups"`
}

func runGroupsHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	const pageSize = 5
	offset := (page - 1) * pageSize

	total, err := storage.GetRunGroupsCount()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to count run groups: %v", err)})
		return
	}

	groups, err := storage.GetRunGroups(pageSize, offset)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load run groups: %v", err)})
		return
	}

	respGroups := make([]runGroupResponse, 0, len(groups))
	for _, g := range groups {
		groupRuns, err := storage.GetRunGroupRuns(g.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to load group runs: %v", err)})
			return
		}

		iters := make([]runGroupRunResponse, 0, len(groupRuns))
		completed := 0
		for _, rr := range groupRuns {
			if rr.Status == "completed" {
				completed++
			}
			row := runGroupRunResponse{
				Iteration: rr.Iteration,
				Score:     rr.Score,
				Weakness:  rr.Weakness,
				Status:    rr.Status,
				Progress:  progressPercent(completed, g.Iterations),
			}
			if rr.DetailJSON.Valid && rr.DetailJSON.String != "" {
				var m models.IterationMetrics
				if err := json.Unmarshal([]byte(rr.DetailJSON.String), &m); err == nil {
					a, sp, sev, st := m.Actionability, m.Specificity, m.Severity, m.Structure
					row.Actionability = &a
					row.Specificity = &sp
					row.Severity = &sev
					row.Structure = &st
					if len(m.Samples) > 0 {
						row.Samples = m.Samples
					}
				}
			}
			iters = append(iters, row)
		}

		respGroups = append(respGroups, runGroupResponse{
			ID:         g.ID,
			InputCode:  g.InputCode,
			Status:     g.Status,
			CreatedAt:  g.CreatedAt,
			UpdatedAt:  g.UpdatedAt,
			Iterations: iters,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(runGroupsResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Groups:   respGroups,
	})
}
