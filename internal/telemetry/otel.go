package telemetry

import "github.com/nkust-monitor-iot-project-2024/central/internal/utils"

// OpenTelemetry module.
//
// telemetry.endpoint.[[metric|trace|log].][grpc|http] - The endpoint to send telemetry data to.
// If not specified, the data will be sent to stdout.
//
// For example,
//
// telemetry.endpoint.metric.http = "localhost:1234"
//							  	will send metrics to the specified HTTP endpoint.
// telemetry.endpoint.grpc = "localhost:2345"
//							   	will send all telemetry data (if not overridden by [metric|trace|log])
// 							   	to the specified gRPC endpoint.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func setupOTelSDK(ctx context.Context, config utils.Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx, config)
	if err != nil {
		handleErr(fmt.Errorf("create trace provider: %w", err))
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, config)
	if err != nil {
		handleErr(fmt.Errorf("create meter provider: %w", err))
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, config)
	if err != nil {
		handleErr(fmt.Errorf("create logger provider: %w", err))
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, config utils.Config) (*trace.TracerProvider, error) {
	traceExporter, err := func() (trace.SpanExporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		for _, k := range []string{"trace.grpc", "grpc"} {
			if v := endpointConfig.String(k); v != "" {
				exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(v))
				if err != nil {
					return nil, fmt.Errorf("set gRPC endpoint %q: %w", v, err)
				}
				return exporter, nil
			}
		}

		for _, k := range []string{"trace.http", "http"} {
			if v := endpointConfig.String(k); v != "" {
				exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(v))
				if err != nil {
					return nil, fmt.Errorf("set HTTP endpoint %q: %w", v, err)
				}
				return exporter, nil
			}
		}

		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("set stdout exporter: %w", err)
		}
		return stdoutExporter, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, config utils.Config) (*metric.MeterProvider, error) {
	metricExporter, err := func() (metric.Exporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		for _, k := range []string{"metric.grpc", "grpc"} {
			if v := endpointConfig.String(k); v != "" {
				exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(v))
				if err != nil {
					return nil, fmt.Errorf("set gRPC endpoint %q: %w", v, err)
				}
				return exporter, nil
			}
		}

		for _, k := range []string{"metric.http", "http"} {
			if v := endpointConfig.String(k); v != "" {
				exporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(v))
				if err != nil {
					return nil, fmt.Errorf("set HTTP endpoint %q: %w", v, err)
				}
				return exporter, nil
			}
		}

		stdoutExporter, err := stdoutmetric.New(
			stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("set stdout exporter: %w", err)
		}
		return stdoutExporter, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, config utils.Config) (*log.LoggerProvider, error) {
	logExporter, err := func() (log.Exporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		for _, k := range []string{"log.http", "http"} {
			if v := endpointConfig.String(k); v != "" {
				exporter, err := otlploghttp.New(ctx, otlploghttp.WithEndpoint(v))
				if err != nil {
					return nil, fmt.Errorf("set HTTP endpoint %q: %w", v, err)
				}
				return exporter, nil
			}
		}

		stdoutExporter, err := stdoutlog.New(
			stdoutlog.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("set stdout exporter: %w", err)
		}
		return stdoutExporter, nil
	}()
	if err != nil {
		return nil, fmt.Errorf("create log exporter: %w", err)
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
