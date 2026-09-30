package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/MacPiggins/gw-analytics/internal/model"
	"github.com/pressly/goose/v3"
)

type Clickhouse struct {
	conn *sql.DB
}

//go:embed migrations/*.sql
var migrations embed.FS

var (
	gooseConfigOnce sync.Once
	gooseConfigErr  error
)

func NewClickhouse(ctx context.Context, conf *config.Config) (*Clickhouse, error) {
	if conf == nil {
		slog.ErrorContext(ctx, "clickhouse config is nil")
		return nil, errors.New("clickhouse config is nil")
	}

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{conf.Clickhouse.Addr},
		Auth: clickhouse.Auth{
			Database: conf.Clickhouse.Database,
			Username: conf.Clickhouse.Username,
			Password: conf.Clickhouse.Password,
		},
	})
	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		slog.ErrorContext(ctx, "error connecting to clickhouse", slog.Any("error", err))
		return nil, err
	}
	
	return &Clickhouse{conn: conn}, nil
}

func (c *Clickhouse) AutoMigrate(ctx context.Context) error {
	if c == nil || c.conn == nil {
		return errors.New("clickhouse connection is nil")
	}
	gooseConfigOnce.Do(func() {
		goose.SetBaseFS(migrations)
		gooseConfigErr = goose.SetDialect("clickhouse")
	})
	if gooseConfigErr != nil {
		return gooseConfigErr
	}
	return goose.UpContext(ctx, c.conn, "migrations")
}

func (c *Clickhouse) Close() error {
	if c == nil || c.conn == nil {
		return errors.New("clickhouse connection is nil")
	}
	return c.conn.Close()
}

func (c *Clickhouse) SaveEvent(ctx context.Context, event *model.Event) error {
	if c == nil || c.conn == nil {
		return errors.New("clickhouse connection is nil")
	}
	if event == nil {
		return errors.New("event is nil")
	}

	_, err := c.conn.ExecContext(ctx, `
		INSERT INTO events (id, type, status, occurred_at, received_at)
		VALUES (?, ?, ?, ?, ?)
	`, event.ID, event.Type, event.Status, event.OccurredAt.UnixMilli(), event.ReceivedAt.UnixMilli())
	return err
}


