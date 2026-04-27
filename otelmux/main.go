package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	serviceName = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure = os.Getenv("INSECURE_MODE")
}

var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE")
)

func main() {
	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	r := mux.NewRouter()

	// Add OpenTelemetry middleware to instrument all routes
	r.Use(otelmux.Middleware("my-gorilla-service"))

	r.HandleFunc("/oke", func(w http.ResponseWriter, r *http.Request) {
		// Extract the span created by otelmux middleware
		span := trace.SpanFromContext(r.Context())

		span.SetAttributes(attribute.String("custom.key", "value"))
		span.AddEvent("manual.event")

		w.Write([]byte("Hello traced Gorilla!"))
	})
	r.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer("my-service")

		// Create a child span (parent created by otelmux middleware)
		_, span := tracer.Start(r.Context(), "process-data")
		defer span.End()

		span.SetAttributes(attribute.String("operation", "data-processing"))

		// Use ctx for downstream calls to propagate trace context

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "processed"}`))
	})
	// log.Println("Server starting on :8080")
	// http.ListenAndServe(":8080", r)
	// srvErr := make(chan error, 1)
	// go func() {
	// 	srvErr <- http.ListenAndServe(":8080", r)
	// }()

	err = http.ListenAndServe(":1323", r)
	logger.Error(err.Error())
}
