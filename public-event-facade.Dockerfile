FROM golang:1.22
WORKDIR /app

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN go install github.com/go-task/task/v3/cmd/task@latest

RUN task service-public-event-facade

EXPOSE 8080
CMD ["/app/out/service-public-event-facade"]
