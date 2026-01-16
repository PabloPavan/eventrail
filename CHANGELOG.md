# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project adheres to Semantic Versioning.

## [Unreleased]

## [0.1.3] - 2026-01-15

### Added
- Comprehensive unit tests for SSE core, publisher, and brokers.
- Redis integration test gated by `INTEGRATION_REDIS_ADDR`.
- GitHub Actions CI job for Redis integration testing.

### Changed
- CI now runs `go test ./...` and runs golangci-lint over `./...`.

## [0.1.2] - 2026-01-15

### Added
- In-memory broker package (`sse/memory`) with `NewBrokerInMemory` constructor.
- README guidance to avoid self-notify in htmx by filtering SSE events via `origin_id`.

### Changed
- `examples/basic` now uses the shared in-memory broker package.
- README includes an in-memory broker alternative for single-process setups.

### Removed
- Example-only in-memory broker implementation from `examples/basic`.

## [0.1.1] - 2026-01-15

### Added
- `NewBrokerPubSub` constructor for Redis Pub/Sub broker.

### Changed
- README example now uses the Redis broker constructor.

## [0.1.0] - 2026-01-15

### Added
- SSE server with hub management, backpressure policies, and lifecycle hooks.
- Broker interface plus Redis Pub/Sub implementation.
- Publisher for JSON events and SSE event encoder defaults.
- `EventNamePrefix` support via encoder wrapping.
- Graceful shutdown integration through `Options.Context` and `Server.Close()`.
- Basic runnable example with in-memory broker (`examples/basic`).
- Linting and vetting helpers (`Makefile`, `.golangci.yml`).
