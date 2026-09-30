package broker

import (
	"context"
	"testing"

	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natscontainer "github.com/testcontainers/testcontainers-go/modules/nats"
)

func testNATS(t *testing.T) (jetstream.Consumer, jetstream.JetStream) {
	t.Helper()
	ctx := context.Background()
	container, err := natscontainer.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		t.Skipf("NATS container unavailable: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	address, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := nats.Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	js, err := jetstream.New(conn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events"},
	}); err != nil {
		t.Fatal(err)
	}
	consumer, err := js.CreateOrUpdateConsumer(ctx, "EVENTS", jetstream.ConsumerConfig{
		Durable: "TEST_CONSUMER",
	})
	if err != nil {
		t.Fatal(err)
	}
	return consumer, js
}

func TestNatsGetMessage(t *testing.T) {
	consumer, js := testNATS(t)
	if _, err := js.Publish(context.Background(), "events", []byte(`{"id":"evt-1","type":"wallet.created","status":"completed","occurred_at":"2026-01-01T00:00:00Z"}`)); err != nil {
		t.Fatal(err)
	}

	br := &Nats{consumer: consumer}
	event, _, err := br.GetMessage()
	if err != nil {
		t.Fatal(err)
	}
	if event == nil || event.ReceivedAt.IsZero() {
		t.Fatal("expected an event with ReceivedAt set")
	}
	if event.ID != "evt-1" {
		t.Fatalf("expected event ID evt-1, got %q", event.ID)
	}
}

func TestNewNatsInitializesConsumer(t *testing.T) {
	ctx := context.Background()
	container, err := natscontainer.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		t.Skipf("NATS container unavailable: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	address, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := nats.Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	js, err := jetstream.New(conn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events"},
	}); err != nil {
		t.Fatal(err)
	}

	br, err := NewNats(ctx, &config.Config{Nats: config.NatsConfig{
		Connstr:  address,
		Stream:   "EVENTS",
		Consumer: "APP_CONSUMER",
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer br.Close()

	if _, err := js.Consumer(ctx, "EVENTS", "APP_CONSUMER"); err != nil {
		t.Fatalf("expected configured durable consumer to exist: %v", err)
	}
}

func TestNatsGetMessageInvalidJSON(t *testing.T) {
	consumer, js := testNATS(t)
	if _, err := js.Publish(context.Background(), "events", []byte("not-json")); err != nil {
		t.Fatal(err)
	}

	br := &Nats{consumer: consumer}
	if _, _, err := br.GetMessage(); err == nil {
		t.Fatal("expected unmarshalling error")
	}
}

func TestNatsGetMessageInvalidEvent(t *testing.T) {
	consumer, js := testNATS(t)
	if _, err := js.Publish(context.Background(), "events", []byte(`{"id":"evt-1"}`)); err != nil {
		t.Fatal(err)
	}

	br := &Nats{consumer: consumer}
	if _, _, err := br.GetMessage(); err == nil {
		t.Fatal("expected validation error for event with missing required fields")
	}
}
