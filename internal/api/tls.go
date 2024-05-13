package api

import (
	"errors"
	"fmt"

	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"google.golang.org/grpc/credentials"
)

func GetTLSCertificateFromConfig(conf utils.Config) (credentials.TransportCredentials, error) {
	caFile := conf.String("server.ca_file")
	certFile := conf.String("server.cert_file")

	if caFile == "" || certFile == "" {
		return nil, errors.New("missing CA or certificate file")
	}

	cred, err := credentials.NewClientTLSFromFile(certFile, caFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS certificate: %w", err)
	}

	return cred, nil
}
