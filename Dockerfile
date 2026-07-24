# syntax=docker/dockerfile:1.7@sha256:a57df69d0ea827fb7266491f2813635de6f17269be881f696fbfdf2d83dda33e

FROM ghcr.io/gohugoio/hugo:v0.164.0@sha256:f8671f2299e60154536c158bff8ce27f6eef4dddbbfc73bcce66263276ae0f80 AS site-builder
WORKDIR /src
COPY --chown=hugo:hugo hugo.toml ./
COPY --chown=hugo:hugo assets ./assets
COPY --chown=hugo:hugo content ./content
COPY --chown=hugo:hugo layouts ./layouts
COPY --chown=hugo:hugo tools ./tools
RUN hugo --destination /tmp/site --gc --minify --noBuildLock

FROM golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS builder
WORKDIR /src

COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY *.go ./
ARG VCS_REPOSITORY_URL
ARG VCS_REVISION
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    LDFLAGS="-s -w"; \
    if [ -n "$VCS_REPOSITORY_URL" ]; then LDFLAGS="$LDFLAGS -X main.vcsRepositoryURL=$VCS_REPOSITORY_URL"; fi; \
    if [ -n "$VCS_REVISION" ]; then LDFLAGS="$LDFLAGS -X main.vcsRevision=$VCS_REVISION"; fi; \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="$LDFLAGS" -o /tooltown .

FROM scratch AS runtime
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder --chown=65532:65532 /tooltown /tooltown
COPY --from=site-builder --chown=65532:65532 /tmp/site ./static

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/tooltown"]
