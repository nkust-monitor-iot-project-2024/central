// Package discover finds the services according to DNS name or configuration and provides them to the caller.
package discover

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/protos/entityrecognitionpb"
	"github.com/samber/mo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Discoverer discovers the services according to DNS name or configuration.
type Discoverer struct {
	config utils.Config

	tracer trace.Tracer
	logger *slog.Logger
}

// NewDiscoverer creates a new Discoverer.
func NewDiscoverer(config utils.Config) *Discoverer {
	const name = "discoverer"

	tracer := otel.GetTracerProvider().Tracer(name)
	logger := utils.NewLogger(name)

	return &Discoverer{config: config, tracer: tracer, logger: logger}
}

// DiscoverEntityRecognitionService discovers the entity recognition service.
//
// It checks if the service URI is set in the configuration "service.entityrecognition.uri",
// and if not, it will use the default value "dns:entity-recognition-service".
//
// The connection is encrypted with the client certificate specified in
// "service.entityrecognition.tls.cert_file" and "service.entityrecognition.tls.key_file".
//
// For more information about the connection details, see Discoverer.CreateGRPCClient.
func (d *Discoverer) DiscoverEntityRecognitionService(ctx context.Context) (entityrecognitionpb.EntityRecognitionClient, error) {
	ctx, span := d.tracer.Start(ctx, "discoverer/DiscoverEntityRecognitionService")
	defer span.End()

	const serviceName = "entityrecognition"
	const fallbackURI = "dns:entity-recognition-service"

	span.AddEvent("create gRPC client")
	client, err := d.CreateGRPCClient(ctx, serviceName, fallbackURI)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create gRPC client")
		span.RecordError(err)

		return nil, fmt.Errorf("create gRPC client: %w", err)
	}
	span.AddEvent("created gRPC client")

	span.AddEvent("create EntityRecognitionClient")
	erClient := entityrecognitionpb.NewEntityRecognitionClient(client)
	span.AddEvent("created EntityRecognitionClient")

	return erClient, nil
}

// CreateGRPCClient creates a gRPC client with the specified URI and TLS certificate.
//
// It checks if the service URI is set in the configuration "service.[serviceName].uri",
// and if not, it will use the default value specified in fallbackURI.
//
// The URI should be specified in the format specified in
// https://github.com/grpc/grpc/blob/master/doc/naming.md#detailed-design.
//
// The connection is encrypted with the client certificate specified in
// "service.[serviceName].tls.cert_file" and "service.[serviceName].tls.key_file".
//
// It DOES NOT check if the service is available or not.
// It is your responsibility to handle the connection error if the service is not available.
//
// The ctx is only used in tracing – closing ctx will not close the connection.
func (d *Discoverer) CreateGRPCClient(ctx context.Context, serviceName string, fallbackURI string) (*grpc.ClientConn, error) {
	ctx, span := d.tracer.Start(ctx, "discoverer/CreateGRPCClient")
	defer span.End()

	serviceConf := d.config.Cut("service." + serviceName)

	span.AddEvent("load TLS certificate")

	certFile := serviceConf.String("tls.cert_file")
	keyFile := serviceConf.String("tls.key_file")
	transportCredentials := mo.None[credentials.TransportCredentials]()

	if certFile != "" && keyFile != "" {
		d.logger.DebugContext(ctx, "found TLS certificate",
			slog.String("certFile", certFile),
			slog.String("keyFile", keyFile))

		cred, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			span.SetStatus(codes.Error, "failed to load TLS credentials")
			span.RecordError(err)

			return nil, fmt.Errorf("load credentials: %w", err)
		}

		transportCredentials = mo.Some(cred)
	} else {
		d.logger.WarnContext(
			ctx,
			"The connection to the service is insecure. Configure the TLS certificate to secure the connection.",
			slog.String("docs", "Discoverer.CreateGRPCClient"),
			slog.String("serviceName", serviceName),
		)
	}
	span.AddEvent("loaded TLS certificate")

	span.AddEvent("create grpc client")
	clientURI := serviceConf.String("uri")
	if clientURI == "" {
		clientURI = fallbackURI
	}

	var dialOptions []grpc.DialOption
	if cert, ok := transportCredentials.Get(); ok {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(cert))
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(clientURI, dialOptions...)
	if err != nil {
		span.SetStatus(codes.Error, "failed to create new client")
		span.RecordError(err)

		return nil, fmt.Errorf("create new client: %w", err)
	}
	span.AddEvent("created grpc client")

	return conn, nil
}
