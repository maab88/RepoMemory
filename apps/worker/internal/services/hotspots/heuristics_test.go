package hotspots

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

func TestDetectProducesDeterministicHotspot(t *testing.T) {
	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	prs := []jobs.PullRequestForHotspot{
		{
			ID:                uuid.New(),
			GitHubPrNumber:    21,
			Title:             "Improve sync retry handling",
			Body:              "retry queue behavior for worker",
			State:             "closed",
			Labels:            []string{"sync", "bug"},
			UpdatedAtExternal: now.Add(-2 * time.Hour),
		},
		{
			ID:                uuid.New(),
			GitHubPrNumber:    20,
			Title:             "Sync pipeline hardening",
			Body:              "worker queue and retry tuning",
			State:             "open",
			Labels:            []string{"sync"},
			UpdatedAtExternal: now.Add(-4 * time.Hour),
		},
	}
	issues := []jobs.IssueForHotspot{
		{
			ID:                uuid.New(),
			GitHubIssueNumber: 99,
			Title:             "Sync failure with retry regression",
			Body:              "queue workers fail intermittently",
			State:             "open",
			Labels:            []string{"sync", "incident"},
			UpdatedAtExternal: now.Add(-1 * time.Hour),
		},
	}

	candidates := Detect(prs, issues)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 hotspot, got %d", len(candidates))
	}
	first := candidates[0]
	if first.Theme != "sync" {
		t.Fatalf("expected sync theme, got %s", first.Theme)
	}
	if first.HitCount != 3 || first.PRCount != 2 || first.IssueCount != 1 {
		t.Fatalf("unexpected counts: %+v", first)
	}
	if len(first.Sources) != 3 {
		t.Fatalf("expected 3 linked sources, got %d", len(first.Sources))
	}
	if first.Sources[0].SourceType != "issue" {
		t.Fatalf("expected most recent issue first, got %s", first.Sources[0].SourceType)
	}
}

func TestDetectLowSignalReturnsNone(t *testing.T) {
	prs := []jobs.PullRequestForHotspot{
		{
			ID:                uuid.New(),
			GitHubPrNumber:    1,
			Title:             "Refactor docs",
			Body:              "small docs tweak",
			State:             "closed",
			Labels:            []string{"docs"},
			UpdatedAtExternal: time.Now().UTC(),
		},
	}
	issues := []jobs.IssueForHotspot{
		{
			ID:                uuid.New(),
			GitHubIssueNumber: 2,
			Title:             "UI typo",
			Body:              "no core subsystem signal",
			State:             "open",
			Labels:            []string{"ui"},
			UpdatedAtExternal: time.Now().UTC(),
		},
	}

	candidates := Detect(prs, issues)
	if len(candidates) != 0 {
		t.Fatalf("expected no hotspots, got %+v", candidates)
	}
}
