package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

/*
* Example of a tail sampling in the otel collector (processor).
* sample all traces if there is an error in any span int the trace, but sample non-error traces using a percentage.
*
tail_sampling:

	decision_wait: 10s
	num_traces: 50000
	expected_new_traces_per_sec: 10
	policies:
		[
			{
				name: error-status-policy,
				type: status_code,
				status_code: { status_codes: [ERROR] },
			},
			{
				name: probabilistic-status-policy,
				type: and,
				and:
					{
						and_sub_policy:
							[
								{
									name: status-policy,
									type: status_code,
									status_code: { status_codes: [OK, UNSET] },
								},
								{
									name: probabilistic-policy,
									type: probabilistic,
									probabilistic: { sampling_percentage: 10 },
								},
							],
					},
			},
		]
*/
func init() {
	err := godotenv.Load(".env")
	if err != nil {
		// Use standard log instead of zap logger to avoid import cycles
		// Logger may not be initialized at this point
		log.Printf("Warning: Failed to load .env file: %v \n", err)
	}
	serviceName = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure = os.Getenv("INSECURE_MODE")
}

var (
	serviceName  string
	collectorURL string
	insecure     string
)

func main() {
	// logger.Debug().Bool("start?", true)
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return err
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Start HTTP server.
	srv := &http.Server{
		Addr:         ":1324",
		BaseContext:  func(net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return err
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return err
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// Register handlers.
	mux.Handle("/rolldice", http.HandlerFunc(rolldice))
	mux.Handle("/rolldice/{player}", http.HandlerFunc(rolldice))
	mux.Handle("/err-rolldice", http.HandlerFunc(errorrolldice))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(
		mux, "/api",
		otelhttp.WithSpanOptions(
			trace.WithAttributes(attribute.Bool("HIHIHI", true)),
		),
	)
	return handler
}
