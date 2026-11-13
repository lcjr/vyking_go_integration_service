package telemetry

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const TracerName = "vyking-integration-service"

// Comment:
// TracingMiddleware injects or extracts OpenTelemetry spans for incoming Partner API requests
func TracingMiddleware(next http.Handler) http.Handler {
	tracer := otel.GetTracerProvider().Tracer(TracerName)
	propagator := otel.GetTextMapPropagator()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, "HTTP "+r.Method+" "+r.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
			semconv.HTTPURL(r.URL.String()),
			semconv.HTTPMethod(r.Method),
		)
		defer span.End()

		// Pass down wrapped request with tracing context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
