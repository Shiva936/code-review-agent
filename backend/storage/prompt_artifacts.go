package storage

import "fmt"

type RunGroupPromptVersion struct {
	ID         int
	RunGroupID int
	Iteration  int
	PromptText string
	RulesJSON  string
	Source     string
	Reason     string
	CreatedAt  string
}

type RunGroupPromptDelta struct {
	ID               int
	RunGroupID       int
	Iteration        int
	WeakestIssue     string
	InputJSON        string
	RawOutput        string
	DeltaJSON        string
	ValidationStatus string
	Applied          bool
	Source           string
	Reason           string
	CreatedAt        string
}

func SaveRunGroupPromptVersion(v *RunGroupPromptVersion) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if v == nil {
		return fmt.Errorf("prompt version is required")
	}
	_, err := db.Exec(
		`INSERT INTO prompt_versions (run_group_id, iteration, prompt_text, rules_json, source, reason)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		v.RunGroupID,
		v.Iteration,
		v.PromptText,
		v.RulesJSON,
		v.Source,
		v.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to save prompt version: %w", err)
	}
	return nil
}

func SaveRunGroupPromptDelta(d *RunGroupPromptDelta) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	if d == nil {
		return fmt.Errorf("prompt delta is required")
	}
	_, err := db.Exec(
		`INSERT INTO prompt_deltas (run_group_id, iteration, weakest_issue, input_json, raw_output, delta_json, validation_status, applied, source, reason)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.RunGroupID,
		d.Iteration,
		d.WeakestIssue,
		d.InputJSON,
		d.RawOutput,
		d.DeltaJSON,
		d.ValidationStatus,
		boolToInt(d.Applied),
		d.Source,
		d.Reason,
	)
	if err != nil {
		return fmt.Errorf("failed to save prompt delta: %w", err)
	}
	return nil
}

func GetRunGroupPromptVersions(groupID int) ([]*RunGroupPromptVersion, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := db.Query(
		`SELECT id, run_group_id, iteration, prompt_text, COALESCE(rules_json, ''), source, COALESCE(reason, ''), created_at
		 FROM prompt_versions
		 WHERE run_group_id = ?
		 ORDER BY iteration ASC, id ASC`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query prompt versions: %w", err)
	}
	defer rows.Close()

	var out []*RunGroupPromptVersion
	for rows.Next() {
		v := &RunGroupPromptVersion{}
		if err := rows.Scan(&v.ID, &v.RunGroupID, &v.Iteration, &v.PromptText, &v.RulesJSON, &v.Source, &v.Reason, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan prompt version: %w", err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prompt versions: %w", err)
	}
	return out, nil
}

func GetRunGroupPromptDeltas(groupID int) ([]*RunGroupPromptDelta, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := db.Query(
		`SELECT id, run_group_id, iteration, weakest_issue, COALESCE(input_json, ''), COALESCE(raw_output, ''), COALESCE(delta_json, ''),
		        validation_status, applied, source, COALESCE(reason, ''), created_at
		 FROM prompt_deltas
		 WHERE run_group_id = ?
		 ORDER BY iteration ASC, id ASC`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query prompt deltas: %w", err)
	}
	defer rows.Close()

	var out []*RunGroupPromptDelta
	for rows.Next() {
		v := &RunGroupPromptDelta{}
		var appliedInt int
		if err := rows.Scan(&v.ID, &v.RunGroupID, &v.Iteration, &v.WeakestIssue, &v.InputJSON, &v.RawOutput, &v.DeltaJSON, &v.ValidationStatus, &appliedInt, &v.Source, &v.Reason, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan prompt delta: %w", err)
		}
		v.Applied = appliedInt == 1
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prompt deltas: %w", err)
	}
	return out, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
