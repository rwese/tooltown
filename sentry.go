package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	otellog "go.opentelemetry.io/otel/log"
)

func initSentry(ctx context.Context) error {
	options := sentryOptionsFromEnvironment()
	if err := sentry.Init(options); err != nil {
		return err
	}
	if options.Dsn != "" {
		sentry.NewMeter(ctx).Count("tooltown.startup", 1)
	}
	return nil
}

func sentryOptionsFromEnvironment() sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Environment: os.Getenv("SENTRY_ENVIRONMENT"),
		Release:     vcsRevision,

		// OpenTelemetry already instruments HTTP requests. Keep Sentry tracing
		// disabled to avoid duplicate transactions and trace propagation.
		EnableTracing: false,
	}
}

func captureSentryLog(ctx context.Context, severity otellog.Severity, message string) {
	logger := sentry.NewLogger(ctx)
	switch {
	case severity >= otellog.SeverityError:
		logger.Error().Emit(message)
	case severity >= otellog.SeverityWarn:
		logger.Warn().Emit(message)
	case severity >= otellog.SeverityInfo:
		logger.Info().Emit(message)
	case severity >= otellog.SeverityDebug:
		logger.Debug().Emit(message)
	default:
		logger.Trace().Emit(message)
	}
}
