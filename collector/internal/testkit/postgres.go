package testkit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	databaseName = "clicky_test"
	databaseUser = "clicky"
	databasePass = "test_password"
)

// TestDatabase exposes two different connection strings deliberately:
// SeedDSN connects straight to PostgreSQL and is only for test setup.
// PgBouncerDSN is the application connection string under test.
type TestDatabase struct {
	SeedDSN      string
	PgBouncerDSN string
	close        func(context.Context) error
}

// NewPostgres starts PostgreSQL and PgBouncer on an isolated Docker network.
// initScripts are applied by PostgreSQL once during its initialization.
func NewPostgres(t testing.TB, initScripts ...string) *TestDatabase {
	t.Helper()

	database, err := NewPostgresForSuite(context.Background(), initScripts...)
	if err != nil {
		t.Fatalf("start test database: %v", err)
	}

	t.Cleanup(func() {
		if err := database.Close(context.Background()); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	return database
}

// NewPostgresForSuite starts PostgreSQL and PgBouncer once for a test package.
// Call Close after m.Run() when using it from TestMain.
func NewPostgresForSuite(ctx context.Context, initScripts ...string) (*TestDatabase, error) {
	network, err := tcnetwork.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create test network: %w", err)
	}

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(databaseName),
		postgres.WithUsername(databaseUser),
		postgres.WithPassword(databasePass),
		postgres.WithInitScripts(initScripts...),
		postgres.BasicWaitStrategies(),
		tcnetwork.WithNetwork([]string{"postgres"}, network),
	)
	if err != nil {
		_ = network.Remove(ctx)
		return nil, fmt.Errorf("start PostgreSQL container: %w", err)
	}

	seedDSN, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		_ = network.Remove(ctx)
		return nil, fmt.Errorf("get PostgreSQL connection string: %w", err)
	}

	pgbouncer, err := testcontainers.Run(
		ctx,
		"edoburu/pgbouncer:latest",
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithEnv(map[string]string{
			"DB_HOST":                 "postgres",
			"DB_PORT":                 "5432",
			"DB_NAME":                 databaseName,
			"DB_USER":                 databaseUser,
			"DB_PASSWORD":             databasePass,
			"AUTH_TYPE":               "plain",
			"POOL_MODE":               "transaction",
			"MAX_CLIENT_CONN":         "20",
			"DEFAULT_POOL_SIZE":       "5",
			"MAX_PREPARED_STATEMENTS": "100",
		}),
		tcnetwork.WithNetwork([]string{"pgbouncer"}, network),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		_ = postgresContainer.Terminate(ctx)
		_ = network.Remove(ctx)
		return nil, fmt.Errorf("start PgBouncer container: %w", err)
	}

	endpoint, err := pgbouncer.PortEndpoint(ctx, "5432/tcp", "")
	if err != nil {
		_ = pgbouncer.Terminate(ctx)
		_ = postgresContainer.Terminate(ctx)
		_ = network.Remove(ctx)
		return nil, fmt.Errorf("get PgBouncer endpoint: %w", err)
	}

	pgbouncerDSN := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&default_query_exec_mode=exec",
		databaseUser,
		databasePass,
		endpoint,
		databaseName,
	)

	if err := waitForPgBouncer(ctx, pgbouncerDSN); err != nil {
		_ = pgbouncer.Terminate(ctx)
		_ = postgresContainer.Terminate(ctx)
		_ = network.Remove(ctx)
		return nil, err
	}

	return &TestDatabase{
		SeedDSN:      seedDSN,
		PgBouncerDSN: pgbouncerDSN,
		close: func(ctx context.Context) error {
			return errors.Join(
				pgbouncer.Terminate(ctx),
				postgresContainer.Terminate(ctx),
				network.Remove(ctx),
			)
		},
	}, nil
}

// Close stops both containers and removes their Docker network.
func (d *TestDatabase) Close(ctx context.Context) error {
	if d.close == nil {
		return nil
	}

	return d.close(ctx)
}

func waitForPgBouncer(ctx context.Context, dsn string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for {
		conn, err := pgx.Connect(ctx, dsn)
		if err == nil {
			_ = conn.Close(ctx)
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for PgBouncer: %w", err)
		case <-time.After(100 * time.Millisecond):
		}
	}
}
