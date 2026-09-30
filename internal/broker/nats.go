package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/MacPiggins/gw-analytics/internal/model"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Nats struct {
	conn     *nats.Conn
	consumer jetstream.Consumer
}

func NewNats(ctx context.Context, conf *config.Config) (*Nats, error) {
	conn, err := nats.Connect(conf.Nats.Connstr)
	if err != nil {
		slog.ErrorContext(ctx, "error connecting to nats server", slog.Any("error", err))
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		slog.ErrorContext(ctx, "error creating jetstream client", slog.Any("error", err))
		return nil, fmt.Errorf("create jetstream client: %w", err)
	}

	consumer, err := js.CreateOrUpdateConsumer(ctx, conf.Nats.Stream, jetstream.ConsumerConfig{
		Durable:       conf.Nats.Consumer,
		FilterSubject: "wallet.events",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		conn.Close()
		slog.ErrorContext(ctx, "error initializing nats consumer", slog.Any("error", err))
		return nil, fmt.Errorf("initialize nats consumer %q on stream %q: %w", conf.Nats.Consumer, conf.Nats.Stream, err)
	}
	return &Nats{conn: conn, consumer: consumer}, nil
}

func (br *Nats) Close() {
	br.conn.Close()
}

func (br *Nats) GetMessage() (*model.Event, func(), error) {
	msg, err := br.consumer.Next()
	if err != nil {
		slog.Error("error retrieving message from nats", slog.Any("error", err))
		return nil, nil, err
	}

	var event model.Event
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		slog.Error("error unmarshalling nats message", slog.Any("error", err))
		return nil, nil, err
	}
	if err := event.Validate(); err != nil {
		slog.Error("invalid nats message", slog.Any("error", err))
		return nil, nil, fmt.Errorf("failed to validate nats message: %w", err)
	}
	event.ReceivedAt = time.Now()
	slog.Info("received nats message")
	return &event, func() {
		if err := msg.Ack(); err != nil {
			slog.Warn("error acknowledging nats message", slog.Any("error", err))
		}
	}, nil
}
