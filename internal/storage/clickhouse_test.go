package storage

import (
	"context"
	"testing"
	"time"

	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/MacPiggins/gw-analytics/internal/model"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func TestClickhouseSaveEvent(t *testing.T) {
	ctx := context.Background()
	container, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:24.3",
		clickhouse.WithDatabase("default"),
		clickhouse.WithUsername("default"),
		clickhouse.WithPassword("default"),
	)
	if err != nil {
		t.Fatalf("clickhouse.Run() error = %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Fatalf("TerminateContainer() error = %v", err)
		}
	})

	addr, err := container.ConnectionHost(ctx)
	if err != nil {
		t.Fatalf("ConnectionHost() error = %v", err)
	}
	store, err := NewClickhouse(ctx, &config.Config{Clickhouse: config.ClickhouseConfig{
		Addr:     addr,
		Database: "default",
		Username: "default",
		Password: "default",
	}})
	if err != nil {
		t.Fatalf("NewClickhouse() error = %v", err)
	}
	defer store.Close()
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}
	if err := store.AutoMigrate(ctx); err != nil {
		t.Fatalf("second AutoMigrate() error = %v", err)
	}

	event := &model.Event{
		ID:         "event-1",
		Type:       "created",
		Status:     "accepted",
		OccurredAt: time.Date(2024, time.January, 2, 3, 4, 5, 123000000, time.UTC),
		ReceivedAt: time.Date(2024, time.January, 2, 3, 4, 6, 456000000, time.UTC),
	}
	if err := store.SaveEvent(ctx, event); err != nil {
		t.Fatalf("SaveEvent() error = %v", err)
	}

	var id, eventType, status string
	var occurredAt, receivedAt time.Time
	if err := store.conn.QueryRowContext(ctx, `
		SELECT id, type, status, occurred_at, received_at
		FROM events WHERE id = ?
	`, event.ID).Scan(&id, &eventType, &status, &occurredAt, &receivedAt); err != nil {
		t.Fatalf("QueryRow().Scan() error = %v", err)
	}
	if event.ID != id {
		t.Fatalf("event ID mismatch: got %q, want %q", id, event.ID)
	}
	if event.Type != eventType {
		t.Fatalf("event type mismatch: got %q, want %q", eventType, event.Type)
	}
	if event.Status != status {
		t.Fatalf("event status mismatch: got %q, want %q", status, event.Status)
	}
	if !event.OccurredAt.Equal(occurredAt) {
		t.Fatalf("occurred_at mismatch: got %v, want %v", occurredAt, event.OccurredAt)
	}
	if !event.ReceivedAt.Equal(receivedAt) {
		t.Fatalf("received_at mismatch: got %v, want %v", receivedAt, event.ReceivedAt)
	}
}
