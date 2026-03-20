package evaluator

// Score represents the evaluation result of a code review.
type Score struct {
	Actionability    int
	Specificity      int
	Severity         int
	Total            int
	WeaknessCategory string
}

// Evaluate scores a generated review using the evaluation rubric.
func Evaluate(review string) (*Score, error) {
	// TODO: Implement evaluation logic
	return &Score{}, nil
}
