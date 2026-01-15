# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project adheres to Semantic Versioning.

## [Unreleased]

## [0.1.0] - 2026-01-15

### Added
- SSE server with hub management, backpressure policies, and lifecycle hooks.
- Broker interface plus Redis Pub/Sub implementation.
- Publisher for JSON events and SSE event encoder defaults.
- `EventNamePrefix` support via encoder wrapping.
- Graceful shutdown integration through `Options.Context` and `Server.Close()`.
- Basic runnable example with in-memory broker (`examples/basic`).
- Linting and vetting helpers (`Makefile`, `.golangci.yml`).
