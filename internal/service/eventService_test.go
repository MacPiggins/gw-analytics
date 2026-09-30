package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MacPiggins/gw-analytics/internal/cache/inmemory"
	"github.com/MacPiggins/gw-analytics/internal/model"
)

type eventStorageMock struct {
	saved *model.Event
	err   error
}

func (m *eventStorageMock) SaveEvent(_ context.Context, event *model.Event) error {
	m.saved = event
	return m.err
}

type eventBrokerMock struct {
	event *model.Event
	err   error
}

func (m *eventBrokerMock) GetMessage() (*model.Event, func(), error) {
	return m.event, nil, m.err
}

func TestEventServiceProcessEvent(t *testing.T) {
	tests := []struct {
		name       string
		brokerErr  error
		storageErr error
	}{
		{name: "broker error", brokerErr: errors.New("get message")},
		{name: "storage error", storageErr: errors.New("save event")},
		{name: "success"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &model.Event{}
			broker := &eventBrokerMock{event: event, err: tt.brokerErr}
			storage := &eventStorageMock{err: tt.storageErr}
			cache := inmemory.NewCache()
			service := NewEventService(storage, broker, cache)

			err := service.processEvent(context.Background())
			wantErr := tt.brokerErr
			if wantErr == nil {
				wantErr = tt.storageErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("processEvent() error = %v, want %v", err, wantErr)
			}
			if tt.brokerErr == nil && storage.saved != event {
				t.Fatalf("saved event = %p, want %p", storage.saved, event)
			}
		})
	}
}

func TestEventServiceStartStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cache := inmemory.NewCache()
	service := NewEventService(&eventStorageMock{}, &eventBrokerMock{}, cache)
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}
}
