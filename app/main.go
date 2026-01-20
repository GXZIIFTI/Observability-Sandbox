package main

import (
	"context"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

)

var logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func main() {
	// Initialize OpenTelemetry
	ctx := context.Background()
	shutdown := initOTel(ctx)
	defer shutdown(ctx)

	// Setup HTTP handlers with automatic tracing
	http.Handle("/healthz", otelhttp.NewHandler(http.HandlerFunc(healthzHandler), "healthz"))
	http.Handle("/work", otelhttp.NewHandler(http.HandlerFunc(workHandler), "work"))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initOTel(ctx context.Context) func(context.Context) {
	// Create resource (identifies this service)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("sample-app"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// Get OTel Collector endpoint
	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "otel-collector:4317"
	}

	// Setup trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelEndpoint),
	)
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
	}

	// Setup trace provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// Setup metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otelEndpoint),
	)
	if err != nil {
		log.Fatalf("failed to create metric exporter: %v", err)
	}

	// Setup metric provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// Return cleanup function
	return func(ctx context.Context) {
		tracerProvider.Shutdown(ctx)
		meterProvider.Shutdown(ctx)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func workHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	traceID := span.SpanContext().TraceID().String()
	log := logger.With(
		"trace_id", traceID,
		"span_id", span.SpanContext().SpanID().String(),
	)

	// Nested span to simulate work
	_, childSpan := otel.Tracer("app").Start(ctx, "simulate_work")
	latency := time.Duration(rand.Intn(400)) * time.Millisecond
	time.Sleep(latency)
	childSpan.End()

	_, cacheSpan := otel.Tracer("app").Start(ctx, "db_cache_lookup")
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	cacheSpan.End()

	// the code fails 20% of the time
	if rand.Float32() < 0.2 {
		log.Error("request failed",
			"latency_ms", latency.Milliseconds(),
			"status", 500,
		)

		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Info("request succeeded",
		"latency_ms", latency.Milliseconds(),
		"status", 200,
	)

	w.Write([]byte("Work completed\n"))
}
