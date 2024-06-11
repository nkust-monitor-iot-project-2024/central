FROM golang:1.22
WORKDIR /app

RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN task service-public-event-facade

EXPOSE 8080
CMD ["/app/out/service-public-event-facade"]
