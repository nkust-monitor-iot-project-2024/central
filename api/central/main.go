package main

import (
	"log/slog"
	"net"

	"github.com/nkust-monitor-iot-project-2024/central/internal/api"
	"github.com/nkust-monitor-iot-project-2024/central/internal/api/central"
	"github.com/nkust-monitor-iot-project-2024/central/internal/database"
	"github.com/nkust-monitor-iot-project-2024/central/internal/utils"
	"github.com/nkust-monitor-iot-project-2024/central/protos/centralpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logger := utils.NewLogger()
	conf := utils.NewConfig(logger)
	db, err := database.ConnectByConfig(conf, logger)
	if err != nil {
		panic(err)
	}

	cert, err := api.GetTLSCertificateFromConfig(conf)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(grpc.Creds(cert))
	service := central.NewService(logger, conf, db)
	centralpb.RegisterCentralServer(server, service)
	reflection.Register(server)

	// Start the server
	listenOn := ":" + conf.String("server.central.port")
	logger.Info("starting central server", slog.String("listen", listenOn))
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
