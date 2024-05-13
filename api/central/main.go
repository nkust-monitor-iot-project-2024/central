package main

import (
	"net"

	"github.com/nkust-monitor-iot-project-2024/central/internal/api/central"
	"github.com/nkust-monitor-iot-project-2024/central/internal/api/common"
	"google.golang.org/grpc"
)

func main() {
	logger := common.NewLogger()
	config := common.NewConfig()

	cert, err := common.GetTLSCertificateFromConfig(config)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(grpc.Creds(cert))
	service := central.NewService(logger, services)
	service.Register(server)

	// Start the server
	net.Listen("tcp", ":10001")
	server.Serve()

}
