package store

import (
	"clicky-go-worker/internal/event"
	"context"
	"encoding/json"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseEventStore struct {
	conn driver.Conn
}

func NewClickHouseEventStore(ctx context.Context, dsn string) (*ClickHouseEventStore, error) {
	conn, err := connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &ClickHouseEventStore{
		conn: conn,
	}, nil
}

func connect(ctx context.Context, dsn string) (driver.Conn, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse ClickHouse DSN: %w", err)
	}

	conn, err := clickhouse.Open(opts)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	return conn, nil
}

func (s *ClickHouseEventStore) Insert(ctx context.Context, events []event.Event) error {
	batch, err := s.conn.PrepareBatch(ctx, `
		INSERT INTO events (
			site_id, token, event_type, url, referrer,
			user_agent, ip, x, y, meta, created_at
		)
	`)
	if err != nil {
		return err
	}

	for _, event := range events {
		meta, err := json.Marshal(event.Meta)
		if err != nil {
			return err
		}

		if err := batch.Append(
			event.SiteID,
			event.Token,
			event.Type,
			event.URL,
			event.Referrer,
			event.UserAgent,
			event.IP.String(),
			event.X,
			event.Y,
			string(meta),
			event.Timestamp,
		); err != nil {
			return err
		}
	}

	return batch.Send()
}
