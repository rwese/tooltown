# Tooltown

A small Go web server for experiments. It currently serves files from `static/`.

## Run

```sh
go run .
```

Open <http://localhost:8080/>.

## Run with Compose

Create local configuration, then build and start the lean container:

```sh
cp .env.dist .env
docker compose up --build
```

Open <http://localhost:8080/>. Changes under `static/` are mounted into the running container. Rebuild after changing Go code.

Use another host port when needed:

```sh
TOOLTOWN_PORT=8081 docker compose up --build
```

Stop it with `docker compose down`.

## OpenTelemetry and Bluebox

The Go service exports request traces, existing application logs, HTTP/runtime metrics, and W3C trace context over OTLP/HTTP. The log exporter preserves standard-log output and timestamps and carries trace context when a log is emitted from a traced context; the current static-file handler emits no request logs. Telemetry stays disabled when `OTEL_EXPORTER_OTLP_ENDPOINT` is empty.

For local Compose, copy `.env.dist` to `.env`, then:

1. Copy `OTEL_EXPORTER_OTLP_ENDPOINT` from `.env.otel.bluebox-template`.
2. Open Bluebox **Setup**, use **Reveal token**, and set the exact displayed value in `OTEL_EXPORTER_OTLP_HEADERS`. Keep the dotenv value quoted because it contains a space; never commit it.
3. Set `OTEL_RESOURCE_ATTRIBUTES=service.namespace=rwese,deployment.environment=development`.

Production deployment reads `OTEL_EXPORTER_OTLP_ENDPOINT` and `OTEL_EXPORTER_OTLP_HEADERS` from secrets in the `tooldown.void.cold.at` GitHub environment and syncs them into `/var/lib/apps/tooltown/.env`; use `deployment.environment=production`. Metrics must keep `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta`; Bluebox rejects cumulative metrics.

## Sentry

Set `SENTRY_DSN` in `.env` from the Sentry project setup and optionally set `SENTRY_ENVIRONMENT`; Compose defaults it to `development` locally and `production` when deployed. Production deployment requires a `SENTRY_DSN` secret in the `tooldown.void.cold.at` GitHub environment.

The service sends application errors, existing application logs, and a `tooltown.startup` count metric to Sentry. Sentry logs and metrics are enabled by default in `sentry-go` v0.48.0. Sentry performance tracing stays disabled because OpenTelemetry already instruments HTTP requests; enabling both would duplicate request transactions and trace propagation.

Published images include `vcs.repository.url.full` and `vcs.ref.head.revision` from CI build metadata so Bluebox can map telemetry to its source revision. Direct `go run .` and local images omit those attributes unless build metadata is supplied.

## Build the image

```sh
docker build --load -t tooltown:dev .
docker run --rm -p 8080:8080 tooltown:dev
```

## Published image

CI runs on the GitHub Actions runner labeled `severed`. Successful builds from `main` are published to GitHub Container Registry as:

```text
ghcr.io/rwese/tooltown:latest
```

Run it with:

```sh
docker run --rm -p 8080:8080 ghcr.io/rwese/tooltown:latest
```

## Deployment

After a successful `main` build, GitHub Actions deploys on the runner labeled `apps-deploy`. It installs `compose.deploy.yaml` as `/var/lib/apps/tooltown/compose.yaml`, creates `.env` from `.env.dist` only when absent, then pulls and starts the configured image.

Traefik discovers the service through Docker labels; the container publishes no host port. Configure the image, hostname, container port, restart policy, Traefik router/service settings, management label, and Watchtower opt-in in the deployment `.env`. All supported variables and defaults are documented in `.env.dist`. Use distinct `COMPOSE_PROJECT_NAME`, `TOOLTOWN_TRAEFIK_ROUTER`, `TOOLTOWN_TRAEFIK_SERVICE`, and `TOOLTOWN_HOST` values for separate deployments on the same host.

The deployment runner requires Docker Compose and write access to `/var/lib/apps`. Subsequent deployments preserve its local `.env`. Each deployment updates `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, and `SENTRY_DSN` from environment secrets of the same names.

## Test

```sh
go test ./...
```
