FROM golang:1.22
WORKDIR /app

RUN apt-get update && apt-get install -y \
    unzip \
    && rm -rf /var/lib/apt/lists/*

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v23.4/protoc-23.4-linux-x86_64.zip && \
    unzip protoc-23.4-linux-x86_64.zip -d /usr/local && \
    rm protoc-23.4-linux-x86_64.zip
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

COPY . .
RUN task service-recognition-facade

CMD ["/app/out/service-recognition-facade"]
