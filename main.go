package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otellog "go.opentelemetry.io/otel/log"
)

const (
	address         = ":8080"
	staticDir       = "static"
	shutdownTimeout = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	telemetry, err := newTelemetry(ctx)
	if err != nil {
		log.Fatal(err)
	}

	telemetry.printf(ctx, otellog.SeverityInfo, "serving %s at http://localhost%s", staticDir, address)
	serveErr := serve(ctx, newHandler(staticDir))
	if serveErr != nil {
		telemetry.printf(ctx, otellog.SeverityError, "%v", serveErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := telemetry.shutdown(shutdownCtx); err != nil {
		log.Printf("shut down OpenTelemetry: %v", err)
	}

	if serveErr != nil {
		os.Exit(1)
	}
}

func serve(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}

func newHandler(dir string) http.Handler {
	return otelhttp.NewHandler(http.FileServer(http.Dir(dir)), "static")
}
