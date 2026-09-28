package portal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNormalizeEmail(t *testing.T) {
	email, err := normalizeEmail("  User@Example.COM ")
	if err != nil || email != "user@example.com" {
		t.Fatalf("normalizeEmail = %q, %v", email, err)
	}
	for _, invalid := range []string{"", "missing-at", "Name <user@example.com>", "a@b.com\nX-Test: yes"} {
		if _, err := normalizeEmail(invalid); err == nil {
			t.Fatalf("normalizeEmail accepted %q", invalid)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	if got := validatePassword("Correct1234"); got != "" {
		t.Fatalf("valid password rejected: %s", got)
	}
	for _, invalid := range []string{"Short1A", "alllowercase1", "ALLUPPERCASE1", "NoDigitsHere"} {
		if got := validatePassword(invalid); got == "" {
			t.Fatalf("weak password accepted: %q", invalid)
		}
	}
}

func TestSecretIsRandomAndHashed(t *testing.T) {
	first, firstHash, err := newSecret(32)
	if err != nil {
		t.Fatal(err)
	}
	second, secondHash, err := newSecret(32)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || firstHash == secondHash {
		t.Fatal("generated secrets must be unique")
	}
	if hashSecret(first) != firstHash || len(firstHash) != 64 || strings.Contains(firstHash, first) {
		t.Fatal("secret hash is invalid")
	}
}

func TestAttemptLimiter(t *testing.T) {
	limiter := newAttemptLimiter()
	if !limiter.Allow("ip", 2, time.Minute) || !limiter.Allow("ip", 2, time.Minute) {
		t.Fatal("first attempts must pass")
	}
	if limiter.Allow("ip", 2, time.Minute) {
		t.Fatal("third attempt must be limited")
	}
	limiter.Reset("ip")
	if !limiter.Allow("ip", 2, time.Minute) {
		t.Fatal("reset must clear attempts")
	}
}

func TestSameOriginAndNoStoreMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{}
	router := gin.New()
	router.POST("/mutation", noStore(), handler.sameOrigin(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	crossSite := httptest.NewRequest(http.MethodPost, "https://axisrelay.example/mutation", nil)
	crossSite.Host = "axisrelay.example"
	crossSite.Header.Set("Origin", "https://attacker.example")
	crossSiteResponse := httptest.NewRecorder()
	router.ServeHTTP(crossSiteResponse, crossSite)
	if crossSiteResponse.Code != http.StatusForbidden {
		t.Fatalf("cross-site response = %d, want %d", crossSiteResponse.Code, http.StatusForbidden)
	}
	if got := crossSiteResponse.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}

	sameSite := httptest.NewRequest(http.MethodPost, "https://axisrelay.example/mutation", nil)
	sameSite.Host = "axisrelay.example"
	sameSite.Header.Set("Origin", "https://axisrelay.example")
	sameSiteResponse := httptest.NewRecorder()
	router.ServeHTTP(sameSiteResponse, sameSite)
	if sameSiteResponse.Code != http.StatusNoContent {
		t.Fatalf("same-site response = %d, want %d", sameSiteResponse.Code, http.StatusNoContent)
	}
}
