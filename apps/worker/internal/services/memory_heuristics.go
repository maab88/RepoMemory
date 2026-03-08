package services

import (
	"regexp"
	"sort"
	"strings"
)

var tokenSplitPattern = regexp.MustCompile(`[^a-z0-9]+`)

func extractImpactedAreas(text string, labels []string) []string {
	areas := map[string]struct{}{}
	combined := strings.ToLower(text + " " + strings.Join(labels, " "))
	tokens := tokenize(combined)

	addIfMatched(areas, tokens, "billing", "billing", "invoice", "payment", "subscription")
	addIfMatched(areas, tokens, "auth", "auth", "oauth", "sso", "token", "login", "permission")
	addIfMatched(areas, tokens, "sync", "sync", "import", "ingest", "backfill")
	addIfMatched(areas, tokens, "queue", "queue", "asynq", "job", "worker", "retry")
	addIfMatched(areas, tokens, "database", "database", "postgres", "sql", "migration", "schema")
	addIfMatched(areas, tokens, "webhooks", "webhook", "webhooks")
	addIfMatched(areas, tokens, "frontend", "frontend", "ui", "nextjs", "react")

	return sortedKeys(areas)
}

func extractRisks(text string, labels []string) []string {
	riskSet := map[string]struct{}{}
	combined := strings.ToLower(text + " " + strings.Join(labels, " "))
	tokens := tokenize(combined)

	if hasAny(tokens, "auth", "oauth", "token", "permission", "permissions") {
		riskSet["Access control behavior changed; verify authorization paths."] = struct{}{}
	}
	if hasAny(tokens, "billing", "invoice", "payment", "subscription") {
		riskSet["Billing-related logic changed; validate financial side effects."] = struct{}{}
	}
	if hasAny(tokens, "migration", "schema", "database", "sql") {
		riskSet["Data migration/schema changes may require rollout validation."] = struct{}{}
	}
	if hasAny(tokens, "queue", "retry", "worker", "sync", "webhook") {
		riskSet["Async processing behavior changed; monitor retries and lag."] = struct{}{}
	}

	return sortedKeys(riskSet)
}

func extractFollowUps(state string, merged bool, text string) []string {
	followUps := map[string]struct{}{}
	stateLower := strings.ToLower(strings.TrimSpace(state))
	textLower := strings.ToLower(text)

	switch {
	case merged:
		followUps["Monitor production behavior after deployment."] = struct{}{}
	case stateLower == "open":
		followUps["Track completion and validate behavior after merge/close."] = struct{}{}
	default:
		followUps["Confirm outcome with stakeholders and archive if complete."] = struct{}{}
	}

	if strings.Contains(textLower, "todo") || strings.Contains(textLower, "follow up") || strings.Contains(textLower, "follow-up") {
		followUps["Capture remaining TODOs as explicit tasks."] = struct{}{}
	}
	if strings.Contains(textLower, "rollback") {
		followUps["Document rollback conditions before release."] = struct{}{}
	}

	return sortedKeys(followUps)
}

func composeWhyItMatters(impactedAreas []string, risks []string, fallback string) string {
	if len(impactedAreas) > 0 {
		return "Touches " + strings.Join(impactedAreas, ", ") + " and may affect downstream workflows."
	}
	if len(risks) > 0 {
		return "Introduces operational risk that should be monitored after rollout."
	}
	return fallback
}

func tokenize(text string) map[string]struct{} {
	parts := tokenSplitPattern.Split(text, -1)
	tokens := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		tokens[p] = struct{}{}
	}
	return tokens
}

func addIfMatched(out map[string]struct{}, tokens map[string]struct{}, area string, keywords ...string) {
	if hasAny(tokens, keywords...) {
		out[area] = struct{}{}
	}
}

func hasAny(tokens map[string]struct{}, keywords ...string) bool {
	for _, kw := range keywords {
		if _, ok := tokens[kw]; ok {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cleanOneLine(text string) string {
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func truncate(text string, max int) string {
	if len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return strings.TrimSpace(text[:max-3]) + "..."
}
