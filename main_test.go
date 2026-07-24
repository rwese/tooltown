package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestSentryOptionsFromEnvironment(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://public@example.invalid/1")
	t.Setenv("SENTRY_ENVIRONMENT", "test")

	previousRevision := vcsRevision
	vcsRevision = "abc123"
	t.Cleanup(func() { vcsRevision = previousRevision })

	options := sentryOptionsFromEnvironment()
	if options.Dsn != "https://public@example.invalid/1" {
		t.Errorf("Dsn = %q", options.Dsn)
	}
	if options.Environment != "test" {
		t.Errorf("Environment = %q", options.Environment)
	}
	if options.Release != "abc123" {
		t.Errorf("Release = %q", options.Release)
	}
	if options.EnableTracing {
		t.Error("EnableTracing = true, want false to avoid duplicating OpenTelemetry HTTP traces")
	}
	if options.DisableLogs {
		t.Error("DisableLogs = true, want Sentry logs enabled")
	}
	if options.DisableMetrics {
		t.Error("DisableMetrics = true, want Sentry metrics enabled")
	}
}

func TestSentryExportsMessageLogAndMetric(t *testing.T) {
	var (
		mu     sync.Mutex
		bodies []string
	)
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read Sentry request: %v", err)
		}
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer receiver.Close()

	dsn := strings.Replace(receiver.URL, "http://", "http://public@", 1) + "/1"
	t.Setenv("SENTRY_DSN", dsn)
	t.Setenv("SENTRY_ENVIRONMENT", "test")
	t.Cleanup(func() { _ = sentry.Init(sentry.ClientOptions{}) })

	if err := initSentry(t.Context()); err != nil {
		t.Fatal(err)
	}
	sentry.CaptureMessage("It works!")
	captureSentryLog(t.Context(), otellog.SeverityInfo, "Sentry log smoke test")
	if !sentry.Flush(2 * time.Second) {
		t.Fatal("Sentry flush timed out")
	}

	mu.Lock()
	payload := strings.Join(bodies, "\n")
	mu.Unlock()
	for _, want := range []string{"It works!", "Sentry log smoke test", "tooltown.startup"} {
		if !strings.Contains(payload, want) {
			t.Errorf("Sentry payload does not contain %q", want)
		}
	}
}

func TestNewResourceWithoutUserEnvironment(t *testing.T) {
	t.Setenv("USER", "")

	if _, err := newResource(t.Context()); err != nil {
		t.Fatalf("newResource() error = %v", err)
	}
}

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
