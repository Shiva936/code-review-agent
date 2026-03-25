package evaluator

import (
	"testing"
)

func TestParseEvaluationResponse_ExtraKeysIgnored(t *testing.T) {
	raw := `Here is JSON:
{
  "actionability": 4,
  "specificity": 3,
  "severity": 5,
  "total": 12,
  "weakness_category": "specificity",
  "logic": 4,
  "performance": 3,
  "security": 5,
  "style": 4,
  "extra_note": "should not break parsing"
}`
	r, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Total != 12 || r.Actionability != 4 || r.WeaknessCategory != "specificity" {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestParseEvaluationResponse_NormalizesWeakness(t *testing.T) {
	raw := `{"actionability":2,"specificity":4,"severity":3,"total":9,"weakness_category":"  SPECIFICITY ","logic":3,"performance":3,"security":3,"style":3}`
	r, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.WeaknessCategory != "specificity" {
		t.Fatalf("want specificity, got %q", r.WeaknessCategory)
	}
}

func TestPickWeakestRubricDim(t *testing.T) {
	if pickWeakestRubricDim(2, 4, 4) != "actionability" {
		t.Fatal()
	}
	if pickWeakestRubricDim(5, 5, 2) != "severity" {
		t.Fatalf("got %s", pickWeakestRubricDim(5, 5, 2))
	}
}
