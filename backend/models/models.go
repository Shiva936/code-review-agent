package models

// CodeReview represents a structured code review output.
type CodeReview struct {
	Comments []Comment
}

// Comment represents a single review comment.
type Comment struct {
	Category string // logic, performance, security, style
	Severity string // critical, minor, suggestion
	Message  string
	Line     int
}

// LoopState represents the persistent state of the improvement loop.
type LoopState struct {
	Iteration       int
	CurrentPrompt   string
	WeaknessHistory map[string]int
	AverageScore    float64
	BestScore       int
}

// Run represents a single iteration result.
type Run struct {
	Iteration int
	Score     int
	Weakness  string
}

// EvalResult represents the evaluation score breakdown.
type EvalResult struct {
	Actionability    int
	Specificity      int
	Severity         int
	Total            int
	WeaknessCategory string
}
