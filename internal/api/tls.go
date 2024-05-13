package api

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"google.golang.org/grpc/credentials"
)

func GetTLSCertificateFromConfig(conf *koanf.Koanf) (credentials.TransportCredentials, error) {
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
