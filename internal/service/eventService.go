package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/MacPiggins/gw-analytics/internal/cache/inmemory"
	"github.com/MacPiggins/gw-analytics/internal/model"
)

type EventStorage interface {
	SaveEvent(ctx context.Context, event *model.Event) error
}

type EventBroker interface {
	GetMessage() (*model.Event, func(), error)
}

type EventService struct {
	storage EventStorage
	broker  EventBroker
	cache   *inmemory.Cache
}

func NewEventService(storage EventStorage, broker EventBroker, cache *inmemory.Cache) *EventService {
	return &EventService{storage: storage, broker: broker, cache: cache}
}

func (svc *EventService) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "event service started")
	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "event service stopped")
			return nil
		default:
			if err := svc.processEvent(ctx); err != nil {
				slog.ErrorContext(ctx, "failed to process event message", slog.Any("error", err))
				continue
			}
		}
	}
}

func (svc *EventService) processEvent(ctx context.Context) error {
	slog.InfoContext(ctx, "waiting for event message")
	event, ack, err := svc.broker.GetMessage()
	if err != nil {
		slog.ErrorContext(ctx, "failed to retrieve event message", slog.Any("error", err))
		return err
	}
	slog.InfoContext(ctx, "received event message", slog.Any("event", event))
	if _, ok := svc.cache.Get(event.ID); !ok {
		svc.cache.Set(event.ID, event, time.Hour*24)
	} else {
		slog.WarnContext(ctx, "duplicate event message", slog.Any("event", event))
		return nil
	}
	if err := svc.storage.SaveEvent(ctx, event); err != nil {
		slog.ErrorContext(ctx, "failed to save event message", slog.Any("error", err))
		return err
	}
	if ack != nil {
		ack()
	}
	return nil
}
