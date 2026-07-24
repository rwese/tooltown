---
name: tooltown
description: "Catalog maintenance for the Tooltown project (rwese/tooltown) — a Go web server that serves a Hugo-generated catalog of small CLI tools. Use when: (1) adding a new tool, (2) editing an existing tool's metadata or screenshots, (3) regenerating the served `static/` output, (4) debugging Hugo, Go, or container builds, (5) understanding the deploy or telemetry pipeline."
license: MIT
---

# Tooltown

A small Go HTTP server that serves a static, Hugo-generated catalog of
CLI tools from `static/`. Source of truth lives in `tools/<slug>/tooltown.yaml`
plus colocated screenshots; Hugo converts each metadata file into
`/tools/<slug>/` via a content adapter.

The live site is at `https://tooltown.void.cold.at/`. The deploy
target is `/var/lib/apps/tooltown/` on the `apps-deploy` runner.

## Layout

```text
main.go, main_test.go, telemetry.go, sentry.go   # Go server
hugo.toml, content/, layouts/, assets/            # Hugo site
tools/<slug>/tooltown.yaml                       # per-tool metadata (canonical)
scripts/build-site                               # pinned Hugo v0.164.0 container build
static/                                          # generated, checked in, served verbatim
docs/catalog-plan.html                           # catalog design + plan
Dockerfile, compose.yaml, compose.deploy.yaml    # container workflow
.github/workflows/container.yaml                 # build, test, publish image (severed)
.github/workflows/deploy.yaml                    # deploy on successful main (apps-deploy)
.env.dist                                        # template for every supported env var
AGENTS.md                                        # project conventions and quality gates
```

## Processes

| Task                         | Where to look                                        |
| ---------------------------- | ---------------------------------------------------- |
| Add a new tool               | `references/new_tool.md`                             |
| Edit existing tool metadata  | `references/new_tool.md` (same schema, same flow)    |
| Refresh metadata from GitHub | Planned (`scripts/fetch-tool` + `.github/workflows/refresh.yaml`); until shipped, edit YAML by hand |
| Regenerate served catalog    | `scripts/build-site` then `git status --short`       |
| Run / build / publish image  | `README.md` (Run, Build the image, Published image)   |
| Deployment                   | `README.md` (Deployment) + `compose.deploy.yaml`      |
| Env vars                     | `.env.dist` (template, full list with comments)      |
| Quality gates                | `AGENTS.md` (test, vet, build-site, container checks)|

## Catalog workflow (one-liner)

Edit YAML → `scripts/build-site` → `git status --short` clean →
`go test ./...` + `go vet ./...` → commit → push → CI runs on
`severed`, deploys on `apps-deploy` if green.

## Hard rules from AGENTS.md

- `static/` is generated. Never edit by hand; CI fails on drift.
- `tooltown.yaml` is canonical. Hugo reads it via the content adapter.
- Conventional Commits. No force-pushes to `main`.
- Update `.env.dist` whenever env vars change.
- Cover behavior changes with tests.

## When this skill does not apply

Skip it for: OpenTelemetry / Sentry / Bluebox setup (use the
`bluebox-otel-instrumentation`, `bluebox-overview`, `production-query`
skills), generic Go or Hugo questions, or anything outside this
repository.
