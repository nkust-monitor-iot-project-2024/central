# IoT Monitor Central Repository

This is a monorepo with the following services:

* [event-aggregator](./cmd/event-aggregator)
* recognition-facade

and the gRPC/Protobuf definition, OpenTelemetry, MQ, database modules.

## Run

This project is highly optimized for [Zeabur](https://zeabur.com),
which is the recommended PaaS for this IoT Monitor project.

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
[telemetry.endpoint.baselime]
api_key = "${BASELIME_API_KEY}"

# SERVICES – Entity Recognition
[services.entityrecognition]
uri = "entity-recognition.your-gpu-machine.internal"
tls = { cert_file = "/etc/iotmonitor/services/entityrecognition/cert.pem", key_file = "/etc/iotmonitor/services/entityrecognition/key.pem" }

# RABBITMQ
[mq]
address = "${RABBITMQ_URI}"
```

Then build the protobuf and `cmd/[service name]/main.go`, and run the binary:

```bash
bash scripts/protoc.sh
go build -o event-aggregator cmd/event-aggregator/main.go 
```

```bash
./event-aggregator
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

## License & Author

This project is licensed under the AGPL-3.0-or-later license. See the [LICENSE](./LICENSE) file for details.

This project is authored by [Yi-Jyun Pan (pan93412)](https://pan93.com).