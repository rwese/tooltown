# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.3] - 2026-07-31

### Added

- Catalog entries for `canonical-clone` and `pdf-decrypt`.
- Tooltown catalog-maintenance skill and new-tool reference.

### Changed

- Focus site copy on tools developed for Tooltown.
- Make each catalog card a single accessible link with hover and keyboard feedback.

### Fixed

- Correct the `canonical-clone` catalog slug.

## [0.3.2] - 2026-07-24

### Fixed

- Make Hugo own copied site sources so restrictive CI checkout permissions cannot block catalog builds.

## [0.3.1] - 2026-07-24

### Fixed

- Use Hugo's pinned multi-platform image index so catalog generation runs on both AMD64 CI and ARM64 development hosts.

## [0.3.0] - 2026-07-24

### Added

- A responsive, accessible project catalog using the selected high-tech monochrome/cyberpunk style.
- Pinned Hugo templates for landing, tool index, tool detail, and about pages.
- Per-tool `tooltown.yaml` metadata and colocated asset structure, seeded with `pwd-copy`.
- Reproducible catalog generation with checked-in output and CI drift detection.

## [0.1.0] - 2026-07-24

### Added

- A Go web server that serves files from the `static/` directory.
- An example static page and automated HTTP tests.
- A lean, non-root multi-stage container image and Compose workflow for local development.
- A GitHub Actions workflow that tests `main`, then builds and publishes its image to GHCR as `latest`.
- A distributable environment template for local Compose configuration.
- Self-hosted CI and automatic deployment of successful `main` builds to `/var/lib/apps/tooltown`.

### Changed

- Made production image, routing, port, restart, management, and update settings configurable through the deployment environment.
