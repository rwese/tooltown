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

The deployment runner requires Docker Compose and write access to `/var/lib/apps`. Subsequent deployments preserve its local `.env`.

## Test

```sh
go test ./...
```
