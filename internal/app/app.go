package app

import (
	"context"

	"github.com/MacPiggins/gw-analytics/internal/broker"
	"github.com/MacPiggins/gw-analytics/internal/cache/inmemory"
	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/MacPiggins/gw-analytics/internal/service"
	"github.com/MacPiggins/gw-analytics/internal/storage"
)

func Run(ctx context.Context, conf *config.Config) error {
	broker, err := broker.NewNats(ctx, conf)
	if err != nil {
		return err
	}

	store, err := storage.NewClickhouse(ctx, conf)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.AutoMigrate(ctx); err != nil {
		return err
	}
	cache := inmemory.NewCache()
	service := service.NewEventService(store, broker, cache)
	if err := service.Start(ctx); err != nil {
		return err
	}
	return nil
}