package ai

import "testing"

func TestParseAndValidateMemorySummary(t *testing.T) {
	valid := `{"type":"pr_summary","title":"Title","summary":"Summary","whyItMatters":"Matters","impactedAreas":["sync"],"risks":["retry"],"followUps":["monitor"]}`
	if _, ok := ParseAndValidateMemorySummary(valid); !ok {
		t.Fatal("expected valid memory summary")
	}

	invalid := `{"type":"pr_summary","title":"Title","summary":"Summary","whyItMatters":"","impactedAreas":[""],"risks":[],"followUps":[]}`
	if _, ok := ParseAndValidateMemorySummary(invalid); ok {
		t.Fatal("expected invalid memory summary")
	}
}

func TestParseAndValidateDigestSummary(t *testing.T) {
	valid := `{"title":"Weekly Digest","summary":"Summary","bodyMarkdown":"## Highlights"}`
	if _, ok := ParseAndValidateDigestSummary(valid); !ok {
		t.Fatal("expected valid digest summary")
	}

	invalid := `{"title":"","summary":"Summary","bodyMarkdown":"body"}`
	if _, ok := ParseAndValidateDigestSummary(invalid); ok {
		t.Fatal("expected invalid digest summary")
	}
}
