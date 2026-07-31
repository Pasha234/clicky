package config

import (
	"testing"
	"time"
)

func TestTokenCacheTTLEnvUsesSharedFormat(t *testing.T) {
	t.Setenv("TOKEN_CACHE_TTL", "30s")

	if got := tokenCacheTTLEnv(); got != 30*time.Second {
		t.Errorf("tokenCacheTTLEnv() = %s, want 30s", got)
	}
}

func TestTokenCacheTTLEnvRejectsMilliseconds(t *testing.T) {
	t.Setenv("TOKEN_CACHE_TTL", "500ms")

	if got := tokenCacheTTLEnv(); got != 5*time.Minute {
		t.Errorf("tokenCacheTTLEnv() = %s, want 5m fallback", got)
	}
}
