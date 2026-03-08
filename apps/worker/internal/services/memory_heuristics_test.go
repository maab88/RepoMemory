package services

import "testing"

func TestExtractImpactedAreas(t *testing.T) {
	areas := extractImpactedAreas("Improve OAuth token retry in worker sync queue", []string{"billing"})
	if len(areas) == 0 {
		t.Fatal("expected impacted areas")
	}
	assertContains(t, areas, "auth")
	assertContains(t, areas, "billing")
	assertContains(t, areas, "queue")
	assertContains(t, areas, "sync")
}

func TestExtractRisksAndFollowUps(t *testing.T) {
	risks := extractRisks("Add migration and permissions checks to webhook queue retry", nil)
	if len(risks) == 0 {
		t.Fatal("expected risks")
	}

	followUps := extractFollowUps("open", false, "todo: follow up rollback plan")
	if len(followUps) == 0 {
		t.Fatal("expected follow-ups")
	}
	assertContains(t, followUps, "Capture remaining TODOs as explicit tasks.")
}

func assertContains(t *testing.T, values []string, target string) {
	t.Helper()
	for _, v := range values {
		if v == target {
			return
		}
	}
	t.Fatalf("expected %q in %v", target, values)
}
