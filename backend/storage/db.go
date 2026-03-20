package storage

// Run represents a single iteration of the loop.
type Run struct {
	ID              int
	Iteration       int
	CodeSnippet     string
	Review          string
	Score           int
	PromptVersion   string
	WeaknessHistory string
}

// InitDB initializes the SQLite database.
func InitDB(path string) error {
	// TODO: Create tables and initialize database
	return nil
}

// SaveRun persists a run result to the database.
func SaveRun(run *Run) error {
	// TODO: Implement persistence logic
	return nil
}

// GetRuns retrieves all runs from the database.
func GetRuns() ([]*Run, error) {
	// TODO: Implement retrieval logic
	return nil, nil
}
