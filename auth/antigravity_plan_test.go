package auth

import (
	"context"
	"testing"
	"time"
)

const antigravityProEntitlementsJSON = `{"allowed":true,"effective_tier":"Google AI Pro",` +
	`"current_tier":{"id":"free-tier","name":"Antigravity"},` +
	`"paid_tier":{"id":"g1-pro-tier","name":"Google AI Pro"},` +
	`"allowed_tiers":[{"id":"free-tier","name":"Antigravity","is_default":true},` +
	`{"id":"standard-tier","name":"Antigravity","user_defined_project":true}],` +
	`"updated_at":"2026-10-20T00:00:00Z"}`

func TestCanonicalAntigravityPlan(t *testing.T) {
	cases := map[string]string{
		"g1-pro-tier":                 "pro",
		"G1-ULTRA-TIER":               "ultra",
		"free-tier":                   "free",
		"standard-tier":               "standard",
		"legacy-tier":                 "legacy",
		"gx-pro-tier":                 "gx-pro",
		"Google AI Pro":               "pro",
		"Google AI Pro (Restricted)":  "pro",
		"Google AI Ultra":             "ultra",
		"Gemini Code Assist Standard": "standard",
		"Antigravity":                 "antigravity",
		"api":                         "api",
		"":                            "",
	}
	for input, want := range cases {
		if got := CanonicalAntigravityPlan(input); got != want {
			t.Errorf("CanonicalAntigravityPlan(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAntigravityEntitlementsPlanKeyUsesTierID(t *testing.T) {
	now := time.Now()
	withName := normalizeAntigravityEntitlements(antigravityLoadProjectResponse{
		CurrentTier: &antigravityTierPayload{ID: "free-tier", Name: "Antigravity"},
		PaidTier:    &antigravityTierPayload{ID: "g1-pro-tier", Name: "Google AI Pro"},
	}, now)
	if withName.EffectiveTier != "Google AI Pro" || withName.PlanKey() != "pro" {
		t.Fatalf("named paid tier: effective=%q key=%q", withName.EffectiveTier, withName.PlanKey())
	}
	withoutName := normalizeAntigravityEntitlements(antigravityLoadProjectResponse{
		PaidTier: &antigravityTierPayload{ID: "g1-pro-tier"},
	}, now)
	if got := withoutName.PlanKey(); got != "pro" {
		t.Fatalf("unnamed paid tier key = %q, want pro", got)
	}
	free := normalizeAntigravityEntitlements(antigravityLoadProjectResponse{
		CurrentTier: &antigravityTierPayload{ID: "free-tier", Name: "Antigravity"},
	}, now)
	if free.EffectiveTier != "Antigravity" || free.PlanKey() != "free" {
		t.Fatalf("free tier: effective=%q key=%q", free.EffectiveTier, free.PlanKey())
	}
	restricted := normalizeAntigravityEntitlements(antigravityLoadProjectResponse{
		AllowedTiers: []antigravityTierPayload{{ID: "free-tier", Name: "Antigravity", IsDefault: true}},
		IneligibleTiers: []struct {
			ReasonCode string `json:"reasonCode"`
		}{{ReasonCode: "REGION"}},
	}, now)
	if restricted.EffectiveTier != "Antigravity (Restricted)" || restricted.PlanKey() != "free" {
		t.Fatalf("restricted tier: effective=%q key=%q", restricted.EffectiveTier, restricted.PlanKey())
	}

	legacy := AntigravityEntitlements{EffectiveTier: "Google AI Ultra (Restricted)"}
	if got := legacy.PlanKey(); got != "ultra" {
		t.Fatalf("legacy display-only key = %q, want ultra", got)
	}

	unknown := AntigravityEntitlements{EffectiveTier: "Google AI Future"}
	if got := unknown.PlanKey(); got != "future" {
		t.Fatalf("unknown tier key = %q, want future", got)
	}
	if got := (AntigravityEntitlements{}).PlanKey(); got != "" {
		t.Fatalf("empty snapshot key = %q, want empty", got)
	}
}

func TestAPIKeyPlanAllowMatchesAntigravityTierKey(t *testing.T) {
	store, db, account, accountID := newAntigravityRefreshTestAccount(t, map[string]any{
		"plan_type":               "Google AI Pro",
		"antigravity_permissions": antigravityProEntitlementsJSON,
	})
	account.mu.RLock()
	planType, planKey := account.PlanType, account.AntigravityPlanKey
	account.mu.RUnlock()
	if planType != "Google AI Pro" || planKey != "pro" {
		t.Fatalf("loaded plan_type=%q key=%q, want display name kept and key pro", planType, planKey)
	}

	store.SetAPIKeyAllowedPlans(1, []string{"pro"})
	store.SetAPIKeyAllowedPlans(2, []string{"ultra"})
	if !store.APIKeyAllowsAccount(1, account) {
		t.Fatal("plan_allow=[pro] must match a Google AI Pro Antigravity account")
	}
	if store.APIKeyAllowsAccount(2, account) {
		t.Fatal("plan_allow=[ultra] must not match a Google AI Pro Antigravity account")
	}

	row, err := db.GetAccountByID(context.Background(), accountID)
	if err != nil {
		t.Fatal(err)
	}
	row.Credentials["plan_type"] = "Google AI Ultra"
	row.Credentials["antigravity_permissions"] = `{"allowed":true,"effective_tier":"Google AI Ultra",` +
		`"paid_tier":{"id":"g1-ultra-tier","name":"Google AI Ultra"},"updated_at":"2026-10-21T00:00:00Z"}`
	generation := store.routingGeneration.Load()
	store.publishAntigravityRuntimeRow(account, row)
	if store.routingGeneration.Load() == generation {
		t.Fatal("plan key change must invalidate routing sub-pools")
	}
	if store.APIKeyAllowsAccount(1, account) || !store.APIKeyAllowsAccount(2, account) {
		t.Fatal("after upgrade, plan_allow must follow the ultra tier key")
	}
}

func TestAPIKeyPlanAllowAntigravityFallsBackToPlanType(t *testing.T) {
	store, _, account, _ := newAntigravityRefreshTestAccount(t, map[string]any{
		"plan_type": "Google AI Pro (Restricted)",
	})
	store.SetAPIKeyAllowedPlans(1, []string{"pro"})
	if !store.APIKeyAllowsAccount(1, account) {
		t.Fatal("rows without an entitlement snapshot must still match via plan_type")
	}
}

func TestAPIKeyPlanAllowAntigravityUnknownPlanFailsClosed(t *testing.T) {
	store, _, account, _ := newAntigravityRefreshTestAccount(t, nil)
	store.SetAPIKeyAllowedPlans(1, []string{"free", "pro", "ultra"})
	if store.APIKeyAllowsAccount(1, account) {
		t.Fatal("an Antigravity account with no known plan must not satisfy a plan whitelist")
	}
	if !store.APIKeyAllowsAccount(2, account) {
		t.Fatal("a key without plan_allow must still reach the account")
	}
}
