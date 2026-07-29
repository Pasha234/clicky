//go:build integration

package token

import (
	"clicky-go-collector/internal/config"
	"clicky-go-collector/internal/testkit"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func requireIntegration(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_INTEGRATION_TESTS=1 to run integration tests")
	}
}

func resetDatabase(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, testDB.SeedDSN)
	if err != nil {
		t.Fatalf("connect for cleanup: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	_, err = conn.Exec(ctx, `TRUNCATE api_tokens, sites CASCADE`)
	if err != nil {
		t.Fatalf("clear test data: %v", err)
	}
}

var testDB *testkit.TestDatabase

func TestMain(m *testing.M) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		os.Exit(m.Run())
	}

	var err error
	testDB, err = testkit.NewPostgresForSuite(
		context.Background(),
		"testdata/schema.sql",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start test database: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	closeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := testDB.Close(closeCtx); err != nil {
		fmt.Fprintf(os.Stderr, "close test database: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func TestPostgresValidatorValidate_ActiveToken(t *testing.T) {
	requireIntegration(t)
	resetDatabase(t)

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, testDB.SeedDSN)
	if err != nil {
		t.Fatalf("connect for seed data: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	const (
		siteID  = "11111111-1111-1111-1111-111111111111"
		tokenID = "22222222-2222-2222-2222-222222222222"
		token   = "clk_active_test_token"
	)

	_, err = conn.Exec(ctx, `
		INSERT INTO sites (id, enabled) VALUES ($1, true);
	`,
		siteID,
	)
	if err != nil {
		t.Fatalf("insert site: %v", err)
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO api_tokens (id, site_id, token) VALUES ($2, $1, $3);
	`,
		siteID,
		tokenID,
		token,
	)
	if err != nil {
		t.Fatalf("insert API token: %v", err)
	}

	validator, err := NewPostgresValidator(ctx, config.Database{URL: testDB.PgBouncerDSN})
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}
	t.Cleanup(validator.Close)

	gotSiteID, err := validator.Validate(ctx, token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if gotSiteID != siteID {
		t.Errorf("Validate() site ID = %q, want %q", gotSiteID, siteID)
	}
}

func TestPostgresValidatorValidate_RevokedToken(t *testing.T) {
	requireIntegration(t)
	resetDatabase(t)

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, testDB.SeedDSN)
	if err != nil {
		t.Fatalf("connect for seed data: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	const (
		siteID  = "11111111-1111-1111-1111-111111111111"
		tokenID = "22222222-2222-2222-2222-222222222222"
		token   = "clk_revoked_test_token"
	)

	_, err = conn.Exec(ctx, `
		INSERT INTO sites (id, enabled) VALUES ($1, true);
	`,
		siteID,
	)
	if err != nil {
		t.Fatalf("insert site: %v", err)
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO api_tokens (id, site_id, token, revoked_at) VALUES ($2, $1, $3, NOW());
	`,
		siteID,
		tokenID,
		token,
	)
	if err != nil {
		t.Fatalf("insert API token: %v", err)
	}

	validator, err := NewPostgresValidator(ctx, config.Database{URL: testDB.PgBouncerDSN})
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}
	t.Cleanup(validator.Close)

	gotSiteID, err := validator.Validate(ctx, token)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want ErrInvalid", err)
	}
	if gotSiteID != "" {
		t.Errorf("Validate() site ID = %v, want empty string", gotSiteID)
	}
}

func TestPostgresValidatorValidate_UnknownToken(t *testing.T) {
	requireIntegration(t)
	resetDatabase(t)

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, testDB.SeedDSN)
	if err != nil {
		t.Fatalf("connect for seed data: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	const (
		siteID = "11111111-1111-1111-1111-111111111111"
		token  = "clk_revoked_test_token"
	)

	_, err = conn.Exec(ctx, `
		INSERT INTO sites (id, enabled) VALUES ($1, true);
	`,
		siteID,
	)
	if err != nil {
		t.Fatalf("insert site: %v", err)
	}

	validator, err := NewPostgresValidator(ctx, config.Database{URL: testDB.PgBouncerDSN})
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}
	t.Cleanup(validator.Close)

	gotSiteID, err := validator.Validate(ctx, token)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Validate() error = %v, want ErrInvalid", err)
	}
	if gotSiteID != "" {
		t.Errorf("Validate() site ID = %v, want empty string", gotSiteID)
	}
}
