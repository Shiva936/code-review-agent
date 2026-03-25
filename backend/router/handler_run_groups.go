package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Shiva936/code-review-agent/backend/config"
	"github.com/Shiva936/code-review-agent/backend/storage"
)

type runGroupRunResponse struct {
	Iteration int    `json:"iteration"`
	Score     int    `json:"score"`
	Weakness  string `json:"weakness"`
}

type runGroupResponse struct {
	ID         int                   `json:"id"`
	Iterations int                   `json:"iterations"`
	CreatedAt  string                `json:"created_at"`
	Runs       []runGroupRunResponse `json:"runs"`
}

type runGroupsResponse struct {
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
	Groups []runGroupResponse `json:"groups"`
}

func runGroupsHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	total, err := storage.GetRunGroupsCount()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("failed to count run groups: %v", err)})
		return
	}

	groups, err := storage.GetRunGroups(limit, offset)
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

		runs := make([]runGroupRunResponse, 0, len(groupRuns))
		for _, rr := range groupRuns {
			runs = append(runs, runGroupRunResponse{
				Iteration: rr.Iteration,
				Score:     rr.Score,
				Weakness:  rr.Weakness,
			})
		}

		respGroups = append(respGroups, runGroupResponse{
			ID:         g.ID,
			Iterations: g.Iterations,
			CreatedAt:  g.CreatedAt,
			Runs:       runs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(runGroupsResponse{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Groups: respGroups,
	})
}
