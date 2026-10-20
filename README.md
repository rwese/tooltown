# Tooltown

A small Go web server for a generated project catalog. It serves files from `static/`.

The catalog uses the selected high-tech interface with switchable monochrome and cyberpunk color schemes. Hugo generates the landing page, tool index, tool detail pages, and about page from checked-in templates and per-tool metadata.

## Add a tool

Create `tools/<slug>/tooltown.yaml` and keep screenshots or other catalog assets in the same directory. Hugo's content adapter turns each metadata file into `/tools/<slug>/` without duplicating content front matter. See `tools/pwd-copy/` for the initial entry.

Generate the checked-in `static/` output with the pinned Hugo v0.164.0 container:

```sh
scripts/build-site
```

Docker is required. CI regenerates the catalog and fails when `static/` has drifted.

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

## OpenTelemetry

The Go service exports request traces, existing application logs, HTTP/runtime metrics, and W3C trace context over OTLP/HTTP. The log exporter preserves standard-log output and timestamps and carries trace context when a log is emitted from a traced context; the current static-file handler emits no request logs. Telemetry stays disabled when `OTEL_EXPORTER_OTLP_ENDPOINT` is empty.

All telemetry goes to the self-hosted OpenTelemetry Collector on Severed at `https://otelcol.void.cold.at`. The collector owns routing, so this service has no vendor endpoint, no ingest token, and no direct egress: traces fan out to Tempo, metrics are scraped through the collector's Prometheus exporter, and logs land in Loki. The route is Tailscale-only behind Traefik's `internal-ipwhitelist` middleware and the collector's OTLP receivers accept unauthenticated traffic, so keep the endpoint on the Tailscale path.

For local Compose, copy `.env.dist` to `.env` and adjust as needed:

1. Leave `OTEL_EXPORTER_OTLP_ENDPOINT` at the collector URL, or set it to a local collector for experiments; an empty value disables telemetry.
2. Set `OTEL_RESOURCE_ATTRIBUTES=service.namespace=rwese,deployment.environment=development`.

The collector converts delta temporality, so `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta` stays required by the Go service.

Production deployment rewrites `OTEL_EXPORTER_OTLP_ENDPOINT` to the collector URL from a non-secret workflow value and syncs only `SENTRY_DSN` from the `tooldown.void.cold.at` GitHub environment into `/var/lib/apps/tooltown/.env`; use `deployment.environment=production`.

## Sentry

Set `SENTRY_DSN` in `.env` from the Sentry project setup and optionally set `SENTRY_ENVIRONMENT`; Compose defaults it to `development` locally and `production` when deployed. Production deployment requires a `SENTRY_DSN` secret in the `tooldown.void.cold.at` GitHub environment.

The service sends application errors, existing application logs, and a `tooltown.startup` count metric to Sentry. Sentry logs and metrics are enabled by default in `sentry-go` v0.48.0. Sentry performance tracing stays disabled because OpenTelemetry already instruments HTTP requests; enabling both would duplicate request transactions and trace propagation.

Published images include `vcs.repository.url.full` and `vcs.ref.head.revision` from CI build metadata so the collector's backends can map telemetry to its source revision. Direct `go run .` and local images omit those attributes unless build metadata is supplied.

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

The deployment runner requires Docker Compose and write access to `/var/lib/apps`. Subsequent deployments preserve its local `.env`, except that each deployment rewrites `OTEL_EXPORTER_OTLP_ENDPOINT` to the collector URL and `SENTRY_DSN` from the environment secret, and drops any leftover `OTEL_EXPORTER_OTLP_HEADERS` line.

## Test

```sh
go test ./...
```
