package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

// external link :
// https://github.com/SigNoz/sample-golang-app/blob/master/main.go
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go
// https://docs.coroot.com/tracing/opentelemetry-go?http-server=echo
var logger zerolog.Logger

func init() {
	logger = zerolog.New(os.Stdout).With().
		Dict("service", zerolog.Dict().
			Str("name", "demo").Str("env", "local")).Logger()

	err := godotenv.Load()
	if err != nil {
		logger.Warn().Msg("Hello from Zerolog global logger")
	}
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
	cleanup := initTracer()
	defer func() {
		_ = cleanup(context.Background())
	}()

	e := echo.New()
	e.Use(otelecho.Middleware(serviceName))

	tracer := otel.Tracer("api")
	e.GET("/", func(c echo.Context) error {
		ctx, span := tracer.Start(c.Request().Context(), "hello-world")
		defer span.End()
		logger.Info().Msg("root for the endpoint")
		go func(ctx context.Context) {
			_, span := tracer.Start(ctx, "ini-child",
				trace.WithAttributes(attribute.String("hello,", " ini child")))
			defer span.End()
			time.Sleep(time.Second)

			_, loopSpan := tracer.Start(ctx, "loop",
				trace.WithAttributes(attribute.Int(
					"jumlah iterasi", 10,
				)))
			for range 10 {
				time.Sleep(200 * time.Millisecond)
			}
			loopSpan.End()
		}(ctx)
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/for-loop", func(c echo.Context) error {
		for i := range 10 {
			_, iSpan := tracer.Start(c.Request().Context(), fmt.Sprintf("Sample-%d", i))
			logger.Info().Msgf("Doing really hard work (%d / 10)\n", i+1)
			if i == 3 || i == 7 {
				time.Sleep(time.Millisecond * 100)
			}
			iSpan.End()
		}
		return c.String(http.StatusOK, "For looping")
	})
	e.Logger.Fatal(e.Start(":1323"))
}

func initTracer() func(context.Context) error {
	var secureOption otlptracegrpc.Option
	if strings.ToLower(insecure) == "false" || insecure == "0" || strings.ToLower("insecure") == "f" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}
	fmt.Println("url: ", collectorURL, serviceName)
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)

	if err != nil {
		logger.Error().Err(err).Msg("failed to get the exporter")
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)

	if err != nil {
		logger.Error().Err(err).Msg("failed to set resources")
	}
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	// Set the propagator
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)
	return exporter.Shutdown
}
