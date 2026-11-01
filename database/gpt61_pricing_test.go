package database

import "testing"

func TestGPT61SolPricingDoesNotUseAstraFallback(t *testing.T) {
	if got := CanonicalBillingModelKey("gpt-6.1-sol-high"); got != "gpt-6.1-sol" {
		t.Fatalf("canonical key = %q, want gpt-6.1-sol", got)
	}
	short := CalculateCostBreakdown(100_000, 10_000, 20_000, "gpt-6.1-sol", "")
	if short.LongContext || short.InputPricePerMToken != 2 || short.CacheReadPricePerMToken != 0.1 || short.OutputPricePerMToken != 10 {
		t.Fatalf("short pricing = %+v", short)
	}
	long := CalculateCostBreakdown(300_000, 10_000, 20_000, "gpt-6.1-sol", "")
	if !long.LongContext || long.InputPricePerMToken != 4 || long.CacheReadPricePerMToken != 0.2 || long.OutputPricePerMToken != 15 {
		t.Fatalf("long pricing = %+v", long)
	}
}
