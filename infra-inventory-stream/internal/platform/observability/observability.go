package observability

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

var (
	AssetsWritten = prometheus.NewCounter(prometheus.CounterOpts{Name: "inventory_assets_written_total", Help: "Assets accepted by the API."})
	EventsReplay  = prometheus.NewCounter(prometheus.CounterOpts{Name: "inventory_events_replayed_total", Help: "Events replayed from the ring buffer."})
	IndexedEvents = prometheus.NewCounter(prometheus.CounterOpts{Name: "inventory_events_indexed_total", Help: "Events written to OpenSearch."})
	IndexFailures = prometheus.NewCounter(prometheus.CounterOpts{Name: "inventory_index_failures_total", Help: "OpenSearch indexing failures."})
)

func Init(ctx context.Context, service string, log *slog.Logger) func(context.Context) error {
	prometheus.MustRegister(AssetsWritten, EventsReplay, IndexedEvents, IndexFailures)
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Warn("trace exporter disabled", "error", err)
		return func(context.Context) error { return nil }
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(service))),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown
}
