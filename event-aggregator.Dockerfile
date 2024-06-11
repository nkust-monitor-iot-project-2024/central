FROM golang:1.22
WORKDIR /app

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN task service-event-aggregator

CMD ["/app/out/service-event-aggregator"]
