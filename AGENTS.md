# Project Guidance

## Approach

- Keep the application small and understandable.
- Prefer Go's standard library; add dependencies only when they clearly reduce complexity.
- Add one behavior at a time. Avoid speculative abstractions and configuration.
- Keep documentation concise and update it when behavior changes.

## Structure

- `main.go`: server entry point and HTTP setup.
- `main_test.go`: HTTP behavior tests.
- `hugo.toml`, `content/`, `layouts/`, `assets/`: Hugo configuration, pages, templates, and authored site styles.
- `tools/<slug>/tooltown.yaml`: canonical tool metadata; keep each tool's screenshots and catalog assets beside it.
- `scripts/build-site`: pinned Hugo container build that replaces `static/`.
- `static/`: generated, checked-in Hugo output served directly by the server; do not edit by hand.
- `docs/catalog-plan.html`: catalog design and implementation plan.
- `Dockerfile`: lean production image built in multiple stages.
- `compose.yaml`: local container workflow with live-mounted static files.
- `compose.deploy.yaml`: production deployment from the published image.
- `.env.dist`: committed template for every supported environment variable.
- `.github/workflows/container.yaml`: tests changes and publishes successful `main` images on the `severed` runner.
- `.github/workflows/deploy.yaml`: deploys successful builds on the `apps-deploy` runner.
- `.workspace/tasks/`: optional, ignored scratch state for active work; delete completed task files.

## Development

- Regenerate the catalog with `scripts/build-site`, then run with `go run .`.
- Format changed Go files with `gofmt`.
- Run `go test ./...`, `go vet ./...`, and `scripts/build-site` before committing; generated `static/` must be clean.
- Validate container changes with `docker compose config`, `docker compose -f compose.deploy.yaml config`, `hadolint Dockerfile`, and `docker build --load .`.
- Keep `.env.dist` updated whenever environment variables are added, changed, or removed.
- Cover behavior changes with tests.

## Git hygiene

- Keep the working tree clean and review `git status --short` before finishing.
- Maintain `.gitignore`: add disposable generated files, local configuration, editor state, and build artifacts as they appear. Catalog output under `static/` is intentionally generated and checked in; never ignore it.
- Do not commit ignored artifacts, secrets, personal data, or machine-specific files.
- Use Conventional Commits.
