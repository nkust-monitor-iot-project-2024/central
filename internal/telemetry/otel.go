// Package telemetry set up OpenTelemetry of a service.
//
// It should be used as soon as possible, so we can trace the service as much as possible.
package telemetry

import (
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/attributext/slogext"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/fx"

	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// FxModule is the fx module for the telemetry that handles the cleanup.
var FxModule = fx.Module(
	"otel",
	fx.Invoke(func(lifecycle fx.Lifecycle, config utils.Config, resource *resource.Resource) error {
		var shutdown OtelShutdownFn
		ctx, cancel := context.WithCancel(context.Background())

		lifecycle.Append(fx.Hook{
			OnStart: func(_ context.Context) (err error) {
				shutdown, err = SetupOTelSDK(ctx, config, resource)
				return
			},
			OnStop: func(_ context.Context) error {
				if shutdown != nil {
					if err := shutdown(ctx); err != nil {
						slog.ErrorContext(ctx, "shutdown telemetry", slogext.Error(err))
						return err
					}
				}

				cancel()

				return nil
			},
		})

		return nil
	}),
)

// OtelShutdownFn is the function returns by SetupOTelSDK that shuts down the OpenTelemetry SDK.
type OtelShutdownFn func(context.Context) error

// SetupOTelSDK sets up the OpenTelemetry SDK.
func SetupOTelSDK(ctx context.Context, config utils.Config, resource *resource.Resource) (shutdown OtelShutdownFn, err error) {
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
	tracerProvider, err := newTraceProvider(ctx, config, resource)
	if err != nil {
		handleErr(fmt.Errorf("create trace provider: %w", err))
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, config, resource)
	if err != nil {
		handleErr(fmt.Errorf("create meter provider: %w", err))
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, config, resource)
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

func newTraceProvider(ctx context.Context, config utils.Config, serviceResource *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := func() (trace.SpanExporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		if otlpEndpoint := endpointConfig.String("otlp.endpoint"); otlpEndpoint != "" {
			options := []otlptracehttp.Option{
				otlptracehttp.WithEndpoint(otlpEndpoint),
			}
			if insecure := endpointConfig.Bool("otlp.insecure"); insecure {
				options = append(options, otlptracehttp.WithInsecure())
			}

			exporter, err := otlptracehttp.New(ctx, options...)
			if err != nil {
				return nil, fmt.Errorf("set OTLP exporter: %w", err)
			}

			return exporter, nil
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
		trace.WithResource(serviceResource),
		trace.WithBatcher(traceExporter),
	)
	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, config utils.Config, serviceResource *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := func() (metric.Exporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		if otlpEndpoint := endpointConfig.String("otlp.endpoint"); otlpEndpoint != "" {
			options := []otlpmetrichttp.Option{
				otlpmetrichttp.WithEndpoint(otlpEndpoint),
			}
			if insecure := endpointConfig.Bool("otlp.insecure"); insecure {
				options = append(options, otlpmetrichttp.WithInsecure())
			}

			exporter, err := otlpmetrichttp.New(ctx, options...)
			if err != nil {
				return nil, fmt.Errorf("set OTLP exporter: %w", err)
			}

			return exporter, nil
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
		metric.WithResource(serviceResource),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, config utils.Config, serviceResource *resource.Resource) (*log.LoggerProvider, error) {
	logExporter, err := func() (log.Exporter, error) {
		endpointConfig := config.Cut("telemetry.endpoint")

		if otlpEndpoint := endpointConfig.String("otlp.endpoint"); otlpEndpoint != "" {
			options := []otlploghttp.Option{
				otlploghttp.WithEndpoint(otlpEndpoint),
			}
			if insecure := endpointConfig.Bool("otlp.insecure"); insecure {
				options = append(options, otlploghttp.WithInsecure())
			}

			exporter, err := otlploghttp.New(ctx, options...)
			if err != nil {
				return nil, fmt.Errorf("set OTLP exporter: %w", err)
			}

			return exporter, nil
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
		log.WithResource(serviceResource),
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
