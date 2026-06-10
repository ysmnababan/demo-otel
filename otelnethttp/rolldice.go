package main

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const name = "go.opentelemetry.io/contrib/examples/dice"

var (
	tracer  = otel.Tracer(name)
	meter   = otel.Meter(name)
	logger  = otelslog.NewLogger(name)
	rollCnt metric.Int64Counter
)

func init() {
	var err error
	rollCnt, err = meter.Int64Counter("dice.rolls",
		metric.WithDescription("The number of rolls by roll value"),
		metric.WithUnit("{roll}"))
	if err != nil {
		panic(err)
	}
}

func rolldice(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	ctx := r.Context()
	defer span.End()

	roll := 1 + rand.Intn(6)

	var msg string
	if player := r.PathValue("player"); player != "" {
		msg = player + " is rolling the dice"
	} else {
		msg = "Anonymous player is rolling the dice"
	}
	logger.InfoContext(ctx, msg, "result", roll)

	rollValueAttr := attribute.Int("roll.value", roll*10)
	span.SetAttributes(rollValueAttr)
	span.SetAttributes(attribute.String("hello", "world"))
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	resp := strconv.Itoa(roll) + "\n"
	if _, err := io.WriteString(w, resp); err != nil {
		logger.ErrorContext(ctx, "Write failed", "error", err)
	}
	span.SetStatus(codes.Ok, "")
	span.SetAttributes(
		semconv.HTTPResponseStatusCode(200),
	)
}

func errorrolldice(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "roll")

	defer span.End()
	roll := 1 + rand.Intn(6)

	var msg string
	if player := r.PathValue("player"); player != "" {
		msg = player + " is rolling the dice"
	} else {
		msg = "Anonymous player is rolling the dice"
	}
	logger.InfoContext(ctx, msg, "result", roll)

	rollValueAttr := attribute.Int("roll.value", roll*10)
	span.SetAttributes(rollValueAttr)
	span.SetAttributes(attribute.String("hello", "world"))
	rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

	resp := strconv.Itoa(roll) + "\n"
	http.Error(w, "internal server error", http.StatusInternalServerError)
	if _, err := io.WriteString(w, resp); err != nil {
		logger.ErrorContext(ctx, "Write failed", "error", err)
	}
	span.SetStatus(codes.Error, "unhandled-error")
	span.RecordError(errors.New("internal server error"))
}
