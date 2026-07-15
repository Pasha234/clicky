package token

import (
	"clicky-go-collector/internal/config"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalid = errors.New("invalid or inactive token")

type Validator interface {
	Validate(ctx context.Context, value string) (siteID string, err error)
}

type Readiness interface {
	Ready(ctx context.Context) error
}

type PostgresValidator struct {
	pool *pgxpool.Pool
}

func NewPostgresValidator(ctx context.Context, cfg config.Database) (*PostgresValidator, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	// Keep the collector below PgBouncer's per-database pool capacity.
	poolConfig.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}

	validator := &PostgresValidator{pool: pool}
	if err := validator.Ready(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return validator, nil
}

func (v *PostgresValidator) Close() {
	v.pool.Close()
}

func (v *PostgresValidator) Ready(ctx context.Context) error {
	if err := v.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return nil
}

func (v *PostgresValidator) Validate(ctx context.Context, value string) (string, error) {
	const query = `
		SELECT sites.id
		FROM api_tokens
		INNER JOIN sites ON sites.id = api_tokens.site_id
		WHERE api_tokens.token = $1
		  AND api_tokens.revoked_at IS NULL
		  AND sites.enabled = true
		LIMIT 1
	`

	var siteID string
	err := v.pool.QueryRow(ctx, query, value).Scan(&siteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalid
	}
	if err != nil {
		return "", fmt.Errorf("validate tracking token: %w", err)
	}

	return siteID, nil
}
