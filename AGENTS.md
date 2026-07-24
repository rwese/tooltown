# Project Guidance

## Approach

- Keep the application small and understandable.
- Prefer Go's standard library; add dependencies only when they clearly reduce complexity.
- Add one behavior at a time. Avoid speculative abstractions and configuration.
- Keep documentation concise and update it when behavior changes.

## Structure

- `main.go`: server entry point and HTTP setup.
- `main_test.go`: HTTP behavior tests.
- `static/`: files served directly by the server.
- `Dockerfile`: lean production image built in multiple stages.
- `compose.yaml`: local container workflow with live-mounted static files.
- `.env.dist`: committed template for every supported environment variable.
- `.github/workflows/container.yaml`: tests changes and publishes successful `main` images.
- `.workspace/tasks/`: optional, ignored scratch state for active work; delete completed task files.

## Development

- Run with `go run .`.
- Format changed Go files with `gofmt`.
- Run `go test ./...` and `go vet ./...` before committing.
- Validate container changes with `docker compose config`, `hadolint Dockerfile`, and `docker build --load .`.
- Keep `.env.dist` updated whenever environment variables are added, changed, or removed.
- Cover behavior changes with tests.

## Git hygiene

- Keep the working tree clean and review `git status --short` before finishing.
- Maintain `.gitignore`: add generated files, local configuration, editor state, and build artifacts as they appear.
- Do not commit ignored artifacts, secrets, personal data, or machine-specific files.
- Use Conventional Commits.
