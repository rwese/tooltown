package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestTelemetryExportsAllSignals(t *testing.T) {
	var (
		mu    sync.Mutex
		paths = make(map[string]bool)
	)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths[r.URL.Path] = true
		mu.Unlock()
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer receiver.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", receiver.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")
	t.Setenv("OTEL_SERVICE_NAME", "tooltown-test")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "deployment.environment=test")

	previousTracerProvider := otel.GetTracerProvider()
	previousMeterProvider := otel.GetMeterProvider()
	previousLoggerProvider := logglobal.GetLoggerProvider()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousTracerProvider)
		otel.SetMeterProvider(previousMeterProvider)
		logglobal.SetLoggerProvider(previousLoggerProvider)
	})

	telemetry, err := newTelemetry(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	telemetry.printf(t.Context(), otellog.SeverityInfo, "telemetry smoke test")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	newHandler(dir).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if err := telemetry.shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range []string{"/v1/traces", "/v1/logs", "/v1/metrics"} {
		if !paths[path] {
			t.Errorf("no export received at %s", path)
		}
	}
}

func TestNewTelemetryRequiresDeltaMetrics(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "https://example.com/otlp")
	t.Setenv("OTEL_SERVICE_NAME", "tooltown")
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "")

	_, err := newTelemetry(t.Context())
	if err == nil || err.Error() != "OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE must be delta" {
		t.Fatalf("error = %v, want delta temporality requirement", err)
	}
}

func TestNewHandlerEmitsServerSpan(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		_ = provider.Shutdown(t.Context())
	})

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	newHandler(dir).ServeHTTP(response, request)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("span count = %d, want 1", len(spans))
	}
	if spans[0].SpanKind != trace.SpanKindServer {
		t.Fatalf("span kind = %v, want %v", spans[0].SpanKind, trace.SpanKindServer)
	}
}

func TestStaticFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	handler := newHandler(dir)

	t.Run("serves index", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if response.Body.String() != "hello" {
			t.Fatalf("body = %q, want %q", response.Body.String(), "hello")
		}
	})

	t.Run("returns not found for missing file", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/missing.txt", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}
