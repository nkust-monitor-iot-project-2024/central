#!/usr/bin/env bash

protoc --proto_path=protos \
    --go_out=. --go_opt=module=github.com/nkust-monitor-iot-project-2024/central \
    --go-grpc_out=. --go-grpc_opt=module=github.com/nkust-monitor-iot-project-2024/central \
    protos/**/*.proto
