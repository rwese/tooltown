package main

import (
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	vcsRepositoryURL string
	vcsRevision      string
)

type telemetry struct {
	logger   otellog.Logger
	shutdown func(context.Context) error
}

func newTelemetry(ctx context.Context) (*telemetry, error) {
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return &telemetry{
			logger:   logglobal.Logger("github.com/rwese/tooltown"),
			shutdown: func(context.Context) error { return nil },
		}, nil
	}
	if os.Getenv("OTEL_SERVICE_NAME") == "" {
		return nil, errors.New("OTEL_SERVICE_NAME must be set when OTEL_EXPORTER_OTLP_ENDPOINT is set")
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE") != "delta" {
		return nil, errors.New("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE must be delta")
	}

	res, err := newResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OpenTelemetry resource: %w", err)
	}

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create OpenTelemetry trace exporter: %w", err)
	}
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		_ = traceProvider.Shutdown(ctx)
		return nil, fmt.Errorf("create OpenTelemetry metric exporter: %w", err)
	}
	metricProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		_ = metricProvider.Shutdown(ctx)
		_ = traceProvider.Shutdown(ctx)
		return nil, fmt.Errorf("create OpenTelemetry log exporter: %w", err)
	}
	logProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetMeterProvider(metricProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	logglobal.SetLoggerProvider(logProvider)

	if err := runtime.Start(runtime.WithMeterProvider(metricProvider)); err != nil {
		_ = logProvider.Shutdown(ctx)
		_ = metricProvider.Shutdown(ctx)
		_ = traceProvider.Shutdown(ctx)
		return nil, fmt.Errorf("start OpenTelemetry runtime metrics: %w", err)
	}

	return &telemetry{
		logger: logProvider.Logger("github.com/rwese/tooltown"),
		shutdown: func(ctx context.Context) error {
			return errors.Join(
				logProvider.Shutdown(ctx),
				metricProvider.Shutdown(ctx),
				traceProvider.Shutdown(ctx),
			)
		},
	}, nil
}

func newResource(ctx context.Context) (*resource.Resource, error) {
	attrs := make([]attribute.KeyValue, 0, 2)
	if vcsRepositoryURL != "" {
		attrs = append(attrs, attribute.String("vcs.repository.url.full", vcsRepositoryURL))
	}
	if vcsRevision != "" {
		attrs = append(attrs, attribute.String("vcs.ref.head.revision", vcsRevision))
	}

	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcessPID(),
		resource.WithProcessExecutableName(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeDescription(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithAttributes(attrs...),
	)
}

func (t *telemetry) printf(ctx context.Context, severity otellog.Severity, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now()

	// Preserve the standard logger's sink, format, timestamp, and precision.
	stdlog.Print(message)
	captureSentryLog(ctx, severity, message)

	var record otellog.Record
	record.SetTimestamp(timestamp)
	record.SetObservedTimestamp(timestamp)
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(otellog.StringValue(message))
	t.logger.Emit(ctx, record)
}
