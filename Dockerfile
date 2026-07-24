# syntax=docker/dockerfile:1.7@sha256:a57df69d0ea827fb7266491f2813635de6f17269be881f696fbfdf2d83dda33e

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

COPY --from=builder --chown=65532:65532 /tooltown /tooltown
COPY --chown=65532:65532 static ./static

USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/tooltown"]
