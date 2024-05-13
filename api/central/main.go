package main

import (
	"net"

	"github.com/nkust-monitor-iot-project-2024/central/internal/api"
	"github.com/nkust-monitor-iot-project-2024/central/internal/api/central"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"google.golang.org/grpc"
)

func main() {
	logger := utils.NewLogger()
	conf := utils.NewConfig()
	db, err := database.ConnectByConfig(conf)
	if err != nil {
		panic(err)
	}

	cert, err := api.GetTLSCertificateFromConfig(conf)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(grpc.Creds(cert))
	service := central.NewService(logger, conf, db)
	service.Register(server)

	// Start the server
	listener, err := net.Listen("tcp", ":"+conf.String("server.central.port"))
	if err != nil {
		panic(err)
	}

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
