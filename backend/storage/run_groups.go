package storage

import (
	"database/sql"
	"fmt"
)

// RunGroup represents a user-triggered run (one input code snippet).
type RunGroup struct {
	ID         int
	InputCode  string
	BasePrompt string
	Iterations int
	CreatedAt  string
}

// RunGroupRun represents a per-iteration result for a group.
type RunGroupRun struct {
	ID        int
	GroupID   int
	Iteration int
	Score     int
	Weakness  string
	CreatedAt string
}

func CreateRunGroup(inputCode string, basePrompt string, iterations int) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	res, err := db.Exec(
		`INSERT INTO run_groups (input_code, base_prompt, iterations) VALUES (?, ?, ?)`,
		inputCode,
		basePrompt,
		iterations,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create run group: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to read run group id: %w", err)
	}

	return int(id), nil
}

func SaveRunGroupRun(groupID int, iteration int, score int, weakness string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		`INSERT INTO run_group_runs (group_id, iteration, score, weakness) VALUES (?, ?, ?, ?)`,
		groupID,
		iteration,
		score,
		weakness,
	)
	if err != nil {
		return fmt.Errorf("failed to save run group run: %w", err)
	}

	return nil
}

// GetRunGroups returns run groups ordered by newest first.
func GetRunGroups(limit int, offset int) ([]*RunGroup, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.Query(
		`SELECT id, input_code, base_prompt, iterations, created_at
		 FROM run_groups
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query run groups: %w", err)
	}
	defer rows.Close()

	var groups []*RunGroup
	for rows.Next() {
		g := &RunGroup{}
		if err := rows.Scan(&g.ID, &g.InputCode, &g.BasePrompt, &g.Iterations, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan run group: %w", err)
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating run groups: %w", err)
	}

	return groups, nil
}

// GetRunGroupRuns returns all runs for a group ordered by iteration.
func GetRunGroupRuns(groupID int) ([]*RunGroupRun, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT id, group_id, iteration, score, weakness, created_at
		 FROM run_group_runs
		 WHERE group_id = ?
		 ORDER BY iteration ASC`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query run group runs: %w", err)
	}
	defer rows.Close()

	var runs []*RunGroupRun
	for rows.Next() {
		rr := &RunGroupRun{}
		if err := rows.Scan(&rr.ID, &rr.GroupID, &rr.Iteration, &rr.Score, &rr.Weakness, &rr.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan run group run: %w", err)
		}
		runs = append(runs, rr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating run group runs: %w", err)
	}

	return runs, nil
}

// GetRunGroupsCount is used for pagination metadata.
func GetRunGroupsCount() (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var count sql.NullInt64
	if err := db.QueryRow(`SELECT COUNT(1) FROM run_groups`).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count run groups: %w", err)
	}
	if !count.Valid {
		return 0, nil
	}
	return int(count.Int64), nil
}
