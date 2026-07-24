# Tooltown

A small Go web server for experiments. It currently serves files from `static/`.

## Run

```sh
go run .
```

Open <http://localhost:8080/>.

## Run with Compose

Build and start the lean container:

```sh
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

## Test

```sh
go test ./...
```
