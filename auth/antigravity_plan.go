package auth

import (
	"encoding/json"
	"strings"

	"github.com/wuekevin/axisrelay/database"
)

// AntigravityPlanAPIKey is the stable plan_allow key for API-key accounts.
const AntigravityPlanAPIKey = "api"

// CanonicalAntigravityPlan maps Google tier IDs and display names to stable
// plan_allow keys. Unknown values remain distinct so existing allowlists fail
// closed instead of treating a new upstream tier as a known plan.
func CanonicalAntigravityPlan(value string) string {
	plan := strings.ToLower(strings.TrimSpace(value))
	plan = strings.TrimSpace(strings.TrimSuffix(plan, "(restricted)"))
	if plan == "" {
		return ""
	}
	if trimmed, ok := strings.CutSuffix(plan, "-tier"); ok {
		if prefix, rest, found := strings.Cut(trimmed, "-"); found && isGoogleOneTierPrefix(prefix) {
			return rest
		}
		return trimmed
	}
	for _, brand := range []string{"google ai ", "gemini code assist "} {
		if rest, ok := strings.CutPrefix(plan, brand); ok {
			return strings.TrimSpace(rest)
		}
	}
	return plan
}

func isGoogleOneTierPrefix(prefix string) bool {
	if len(prefix) < 2 || prefix[0] != 'g' {
		return false
	}
	for _, r := range prefix[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func effectiveAntigravityTier(e *AntigravityEntitlements) (tier *AntigravityTier, fromAllowed bool) {
	if e.PaidTier != nil && firstNonEmpty(e.PaidTier.Name, e.PaidTier.ID) != "" {
		return e.PaidTier, false
	}
	if !e.Restricted && e.CurrentTier != nil && firstNonEmpty(e.CurrentTier.Name, e.CurrentTier.ID) != "" {
		return e.CurrentTier, false
	}
	if len(e.AllowedTiers) == 0 {
		return nil, false
	}
	selected := e.AllowedTiers[0]
	for _, candidate := range e.AllowedTiers {
		if candidate.IsDefault {
			selected = candidate
			break
		}
	}
	if firstNonEmpty(selected.Name, selected.ID) == "" {
		return nil, false
	}
	return &selected, true
}

// PlanKey returns the canonical authorization key while EffectiveTier keeps
// the human-readable Google label used by the admin UI.
func (e AntigravityEntitlements) PlanKey() string {
	tier, _ := effectiveAntigravityTier(&e)
	if tier == nil {
		return CanonicalAntigravityPlan(e.EffectiveTier)
	}
	return CanonicalAntigravityPlan(firstNonEmpty(tier.ID, tier.Name))
}

func antigravityPlanKeyFromRow(row *database.AccountRow) string {
	if row == nil {
		return ""
	}
	if strings.TrimSpace(row.GetCredential("api_key")) != "" {
		return AntigravityPlanAPIKey
	}
	raw := strings.TrimSpace(row.GetCredential("antigravity_permissions"))
	if raw == "" {
		raw = strings.TrimSpace(row.GetCredential("antigravity_entitlements"))
	}
	if raw != "" {
		var entitlements AntigravityEntitlements
		if json.Unmarshal([]byte(raw), &entitlements) == nil {
			if key := entitlements.PlanKey(); key != "" {
				return key
			}
		}
	}
	return CanonicalAntigravityPlan(row.GetCredential("plan_type"))
}
