package storage

import "fmt"

type RunGroupPromptVersion struct {
	RunGroupID int
	Iteration  int
	PromptText string
	RulesJSON  string
	Source     string
	Reason     string
}

type RunGroupPromptDelta struct {
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
