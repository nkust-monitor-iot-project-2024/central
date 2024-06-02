# IoT Monitor Central Repository

This is a monorepo with the following services:

* [event-aggregator](./cmd/event-aggregator)
* [recognition-facade](./cmd/recognition-facade)
* public-event-facade

and the gRPC/Protobuf definition, OpenTelemetry, MQ, database modules.

## Run

This project is highly optimized for [Zeabur](https://zeabur.com),
which is the recommended PaaS for this IoT Monitor project.

You should prepare your PostgreSQL, RabbitMQ and [EntityRecognition](https://github.com/nkust-monitor-iot-project-2024/recognition) instance.
Besides, you may need to configure [Grafana Alloy](https://grafana.com/oss/alloy-opentelemetry-collector/),
so you can monitor the telemetry of services.

Place the following config.toml in the working directory or `/etc/iotmonitor/config.toml`, `~/.config/iotmonitor/config.toml`:

```toml
# DATABASE
[postgres]
host = "${POSTGRES_HOST}"
port = "5432"  # or leave blank for default
user = "${POSTGRES_USER}"
password = "${POSTGRES_PASSWORD}"
dbname = "${POSTGRES_DB}"
sslmode = "disable"  # https://www.postgresql.org/docs/current/libpq-ssl.html

# TELEMETRY
[telemetry.endpoint.otlp]
endpoint = "localhost:4318"
insecure = true

# SERVICES – Entity Recognition
[service.entityrecognition]
uri = "entity-recognition.your-gpu-machine.internal"
tls = { cert_file = "/etc/iotmonitor/services/entityrecognition/cert.pem", key_file = "/etc/iotmonitor/services/entityrecognition/key.pem" }

# SERVICES - Public Event Facade
[service.publiceventfacade]
port = 1145  # by default, it is 8080
[service.publiceventfacade.tls]
cert_file = "/etc/iotmonitor/services/publiceventfacade/cert.pem"
key_file = "/etc/iotmonitor/services/publiceventfacade/key.pem"

# RABBITMQ
[mq]
address = "${RABBITMQ_URI}"
```

To build the service, you should install [Taskfile runner](https://taskfile.dev/usage/), then you can run the command to build the microservices:

```bash
task service-event-aggregator  # task service-[service name], single service
task service-all               # all services
```

Then you can run the service:

```bash
./out/event-aggregator
```

## Development

You should build the protobuf first.

```bash
bash scripts/protoc.sh
```

I use [GoLand](https://www.jetbrains.com/go/) for development, but you can use any editor you like.

If you change the ent schema, you should run the following command to update the schema:

```bash
go generate ./ent
```

## Production deployment

wip (Zeabur template)

## Specification

### AMQP Message Specification

For a `Event`:

- The Message ID is the event ID, and must be in UUID format.
- The Timestamp will be the EmittedAt of the event.
- The App ID should be your device ID.
- The Content Type should be `application/x-google-protobuf`, and the Type should be `eventpb.EventMessage`.
- The body should be formed in [`eventpb.EventMessage`](protos/eventpb/event.proto) and encoded in Protobuf Binary format.
- The header can contain the W3C Trace Context and W3C Baggage information.
  - The header key of Trace Context should be `tracestate` and `traceparent`.
  - The header key of Baggage should be `baggage`.
- The header can contain the `parent_event_id` optionally. If there is one, it must be in UUID format.

The implementation can be seen in [publish_event.go](internal/mq/publish_event.go).

## License & Author

This project is licensed under the AGPL-3.0-or-later license. See the [LICENSE](./LICENSE) file for details.

This project is authored by [Yi-Jyun Pan (pan93412)](https://pan93.com).