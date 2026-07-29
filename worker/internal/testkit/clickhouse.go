//go:build integration

package testkit

import (
	"context"
	"fmt"
	"testing"
	"time"

	clickhouse2 "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

const (
	clickhouseUser = "clicky"
	clickhousePass = "test_password"
	clickhouseDB   = "clicky"
)

type TestClickHouse struct {
	DSN   string
	close func(context.Context) error
}

func NewClickHouse(t testing.TB, initScripts ...string) *TestClickHouse {
	t.Helper()

	db, err := NewClickHouseForSuite(context.Background(), initScripts...)
	if err != nil {
		t.Fatalf("start test ClickHouse: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(context.Background()); err != nil {
			t.Errorf("close test ClickHouse: %v", err)
		}
	})

	return db
}

func NewClickHouseForSuite(ctx context.Context, initScripts ...string) (*TestClickHouse, error) {
	container, err := clickhouse.Run(
		ctx,
		"clickhouse/clickhouse-server:24.8",
		clickhouse.WithUsername(clickhouseUser),
		clickhouse.WithPassword(clickhousePass),
		clickhouse.WithDatabase(clickhouseDB),
		clickhouse.WithInitScripts(initScripts...),
	)

	if err != nil {
		return nil, fmt.Errorf("start clickhouse container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("get ClickHouse connection string: %w", err)
	}

	if err := waitForClickHouse(ctx, dsn); err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}

	return &TestClickHouse{
		DSN: dsn,
		close: func(ctx context.Context) error {
			return container.Terminate(ctx)
		},
	}, nil
}

func (b *TestClickHouse) Close(ctx context.Context) error {
	if b.close == nil {
		return nil
	}

	return b.close(ctx)
}

func waitForClickHouse(ctx context.Context, dsn string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for {
		conn, err := ClickHouseConnect(ctx, dsn)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for ClickHouse: %w", err)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func ClickHouseConnect(ctx context.Context, dsn string) (driver.Conn, error) {
	opts, err := clickhouse2.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse ClickHouse DSN: %w", err)
	}

	conn, err := clickhouse2.Open(opts)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		if exception, ok := err.(*clickhouse2.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	return conn, nil
}
