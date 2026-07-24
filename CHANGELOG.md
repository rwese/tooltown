# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- A Go web server that serves files from the `static/` directory.
- An example static page and automated HTTP tests.
- A lean, non-root multi-stage container image and Compose workflow for local development.
- A GitHub Actions workflow that tests `main`, then builds and publishes its image to GHCR as `latest`.
