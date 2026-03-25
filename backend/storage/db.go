package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Shiva936/code-review-agent/backend/config"
	_ "modernc.org/sqlite"
)

// DB holds the database connection
var db *sql.DB

// Run represents a single iteration result.
type Run struct {
	Iteration int
	Score     int
	Weakness  string
}

// PromptVersion represents a stored prompt version.
type PromptVersion struct {
	Version    int
	PromptText string
	Reason     string
}

// InitDB initializes the SQLite database and creates tables if they don't exist.
func InitDB(cfg *config.Config) error {
	if cfg.DatabasePath == "" {
		// Use environment variable or default path
		if envPath := os.Getenv("DATABASE_PATH"); envPath != "" {
			cfg.DatabasePath = envPath
		} else {
			// In deployments (e.g. Docker / Railway with a mounted volume),
			// prefer the mounted path.
			if runtime.GOOS == "windows" {
				cfg.DatabasePath = "./data/app.db"
			} else {
				cfg.DatabasePath = "/data/app.db"
			}
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// If a previous connection exists (e.g. tests), close it before reopening.
	if db != nil {
		_ = db.Close()
		db = nil
	}

	var err error
	db, err = sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	// Create runs table
	runsTable := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		iteration INTEGER NOT NULL,
		score INTEGER NOT NULL,
		weakness TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(runsTable); err != nil {
		return fmt.Errorf("failed to create runs table: %w", err)
	}

	// Create prompts table
	promptsTable := `
	CREATE TABLE IF NOT EXISTS prompts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		version INTEGER NOT NULL UNIQUE,
		prompt_text TEXT NOT NULL,
		reason TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(promptsTable); err != nil {
		return fmt.Errorf("failed to create prompts table: %w", err)
	}

	// Create run_groups table (grouped run storage)
	runGroupsTable := `
	CREATE TABLE IF NOT EXISTS run_groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		input_code TEXT NOT NULL,
		base_prompt TEXT NOT NULL,
		iterations INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(runGroupsTable); err != nil {
		return fmt.Errorf("failed to create run_groups table: %w", err)
	}

	// Create run_group_runs table (per-iteration results for a group)
	runGroupRunsTable := `
	CREATE TABLE IF NOT EXISTS run_group_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		iteration INTEGER NOT NULL,
		score INTEGER NOT NULL,
		weakness TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(group_id) REFERENCES run_groups(id)
	)`

	if _, err := db.Exec(runGroupRunsTable); err != nil {
		return fmt.Errorf("failed to create run_group_runs table: %w", err)
	}

	return nil
}

// SaveRun persists a run result to the database.
func SaveRun(run *Run) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `INSERT INTO runs (iteration, score, weakness) VALUES (?, ?, ?)`
	_, err := db.Exec(query, run.Iteration, run.Score, run.Weakness)
	if err != nil {
		return fmt.Errorf("failed to save run: %w", err)
	}

	return nil
}

// GetRuns retrieves all runs from the database ordered by iteration.
func GetRuns() ([]*Run, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT iteration, score, weakness FROM runs ORDER BY iteration ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		run := &Run{}
		if err := rows.Scan(&run.Iteration, &run.Score, &run.Weakness); err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}
		runs = append(runs, run)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating runs: %w", err)
	}

	return runs, nil
}

// SavePromptVersion saves a new prompt version to the database.
func SavePromptVersion(version int, promptText string, reason string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `INSERT INTO prompts (version, prompt_text, reason) VALUES (?, ?, ?)`
	_, err := db.Exec(query, version, promptText, reason)
	if err != nil {
		return fmt.Errorf("failed to save prompt version: %w", err)
	}

	return nil
}

// GetPromptVersions retrieves all prompt versions from the database.
func GetPromptVersions() ([]*PromptVersion, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT version, prompt_text, reason FROM prompts ORDER BY version ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query prompt versions: %w", err)
	}
	defer rows.Close()

	var prompts []*PromptVersion
	for rows.Next() {
		prompt := &PromptVersion{}
		if err := rows.Scan(&prompt.Version, &prompt.PromptText, &prompt.Reason); err != nil {
			return nil, fmt.Errorf("failed to scan prompt version: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating prompt versions: %w", err)
	}

	return prompts, nil
}

// Close closes the database connection
func Close() error {
	if db != nil {
		err := db.Close()
		db = nil
		return err
	}
	return nil
}
