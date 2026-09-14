package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CaiqueGOliveira/TelemetryGo/src/bootstrap"
)

func main() {
	app, err := bootstrap.AppBootstrap()
	if err != nil {
		slog.Error("failed to bootstrap application", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":" + app.ServerPort,
		Handler: app.Router,
	}

	go func() {
		slog.Info("server running", "port", app.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	app.Cleanup()
	slog.Info("server stopped")
}
