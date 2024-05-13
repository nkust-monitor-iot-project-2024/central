package api

import (
	"errors"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"google.golang.org/grpc/credentials"
)

func GetTLSCertificateFromConfig(conf utils.Config) (credentials.TransportCredentials, error) {
	certFile := conf.String("server.cert_file")
	keyFile := conf.String("server.key_file")

	if certFile == "" || keyFile == "" {
		return nil, errors.New("missing cert or key file")
	}

	cred, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS certificate: %w", err)
	}

	return cred, nil
}
