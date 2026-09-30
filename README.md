# GW Analytics

GW Analytics is a Go service that consumes wallet events from NATS JetStream and stores them in ClickHouse for downstream analytics and reporting. It is designed to be lightweight, container-friendly, and easy to run locally or in Docker.

## Overview

The application:

- connects to a NATS server and creates or updates a JetStream consumer
- listens on the `wallet.events` subject
- deserializes incoming event payloads into the internal `Event` model
- stores events in ClickHouse with schema auto-migration
- keeps processing until the application context is canceled

The service is intentionally simple: it reads messages from NATS, saves them, and logs the result. It does not perform heavy transformations or expose an HTTP API.

## Architecture

- `cmd/main.go` starts the app and config loading
- `internal/app/app.go` wires the broker, store, and event service together
- `internal/broker/nats.go` reads JetStream messages from NATS
- `internal/storage/clickhouse.go` connects to ClickHouse and persists events
- `internal/storage/migrations/00001_create_events.sql` creates the analytics table
- `internal/service/eventService.go` processes messages in a loop

## Event model

Each event saved by the service contains:

| Field | Type | Description |
| --- | --- | --- |
| `id` | `String` | Unique event identifier |
| `type` | `String` | Event type |
| `status` | `String` | Event status |
| `occurred_at` | `DateTime64(3)` | When the event occurred |
| `received_at` | `DateTime64(3)` | When the service received it |

## Configuration

The app loads configuration from `config.env` by default using `godotenv`.

Copy the example file and adjust values for your environment:

```bash
cp config.env.example config.env
```

Supported environment variables:

```env
NATSCONNSTR=nats://localhost:4222
NATSSTREAM=analytics
NATSCONSUMER=analytics

CLICKHOUSEADDR=localhost:9000
CLICKHOUSEDB=default
CLICKHOUSEUSER=default
CLICKHOUSEPASSWORD=changeme
```

Defaults in code are:

- NATS address: `nats://localhost:4222`
- Stream: `analytics`
- Consumer: `analytics`
- ClickHouse address: `localhost:9000`
- Database: `default`
- User: `default`

## Running locally

### Option 1: Docker Compose

This project includes a Docker Compose setup that starts NATS and ClickHouse, then runs the analytics service:

```bash
docker compose up --build
```

This uses the service defined in `docker-compose.yaml` with dependencies on both NATS and ClickHouse.

### Option 2: Run the Go binary directly

Start the required infrastructure manually, then run:

```bash
go run ./cmd/main.go
```

If you are running NATS and ClickHouse on localhost, the default config settings should work without additional changes.

## Docker image

The application is built into a minimal binary image using the included `Dockerfile`:

```bash
docker build -t gw-analytics .
```

## Notes

- The service expects a JetStream stream with the `wallet` name.
- It listens to the `wallet.events` subject only.
- If the service cannot read or save an event, it logs the error and continues processing.
- ClickHouse schema is migrated automatically on startup with Goose.
