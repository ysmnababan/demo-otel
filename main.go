package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// external link :
// https://github.com/SigNoz/sample-golang-app/blob/master/main.go
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go
// https://docs.coroot.com/tracing/opentelemetry-go?http-server=echo
func init() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
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

	trace := otel.Tracer("api")
	e.GET("/", func(c echo.Context) error {
		_, span := trace.Start(c.Request().Context(), "hello-world")
		defer span.End()

		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/for-loop", func(c echo.Context) error {
		for i := range 10 {
			_, iSpan := trace.Start(c.Request().Context(), fmt.Sprintf("Sample-%d", i))
			log.Printf("Doing really hard work (%d / 10)\n", i+1)
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
		log.Fatalf("failed to create exporter: %v", err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)

	if err != nil {
		log.Fatalf("failed to set resources: %v", err)
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
