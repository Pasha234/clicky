package config

import (
	"testing"
	"time"
)

func TestDurationEnv(t *testing.T) {
	t.Setenv("FLUSH_INTERVAL", "2s")

	got, err := durationEnv("FLUSH_INTERVAL", 10*time.Second)
	if err != nil {
		t.Fatalf("durationEnv() error = %v", err)
	}

	if got != 2*time.Second {
		t.Fatalf("durationEnv() = %s, want 2s", got)
	}
}

func TestDurationEnvRejectInvalidValue(t *testing.T) {
	t.Setenv("FLUSH_INTERVAL", "not-a-duration")

	_, err := durationEnv("FLUSH_INTERVAL", 10*time.Second)
	if err == nil {
		t.Fatalf("durationEnv() error = nil, want error")
	}
}

func TestIntEnv(t *testing.T) {
	t.Setenv("BATCH_SIZE", "1")

	got, err := intEnv("BATCH_SIZE", 10)
	if err != nil {
		t.Fatalf("intEnv() error = %v", err)
	}

	if got != 1 {
		t.Fatalf("intEnv() = %d, want 1", got)
	}
}

func TestIntEnvRejectInvalidValue(t *testing.T) {
	t.Setenv("BATCH_SIZE", "not-a-int")

	_, err := intEnv("BATCH_SIZE", 10)
	if err == nil {
		t.Fatalf("intEnv() error = nil, want error")
	}
}

func TestIntEnvRejectsZero(t *testing.T) {
	t.Setenv("BATCH_SIZE", "0")
	_, err := intEnv("BATCH_SIZE", 10)
	if err == nil {
		t.Fatalf("intEnv() error = nil, want error")
	}
}

func TestIntEnvRejectsNegative(t *testing.T) {
	t.Setenv("BATCH_SIZE", "-1")
	_, err := intEnv("BATCH_SIZE", 10)
	if err == nil {
		t.Fatalf("intEnv() error = nil, want error")
	}
}
