package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/MacPiggins/gw-analytics/internal/app"
	"github.com/MacPiggins/gw-analytics/internal/config"
	"github.com/MacPiggins/gw-analytics/internal/logging"
)

func main() {
	h := &logging.ContextHandler{Handler: slog.NewJSONHandler(os.Stdout, nil)}
	slog.SetDefault(slog.New(h))

	conf, err := config.Load("config.env")
	if err != nil {
		slog.Error("error while loading config", slog.Any("error", err))
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = app.Run(ctx, conf)
	if err != nil {
		slog.Error("error running analytics service", slog.Any("error", err))
	}
}
