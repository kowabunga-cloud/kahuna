# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Kahuna** is the orchestration API gateway for the Kowabunga cloud platform. It is a Go REST API server that manages cloud resources (compute, storage, networking) and persists state in MongoDB.

## Build & Development Commands

```bash
make all       # Full pipeline: mod + fmt + vet + lint + build
make build     # Compile binary to ./bin/
make mod       # Download and tidy Go dependencies
make fmt       # Format code with gofmt
make vet       # Run go vet
make lint      # Run golangci-lint
make tests     # Run all tests with coverage (go test ./... -count=1 -coverprofile=coverage.txt)
make sec       # Run gosec security checks
make vuln      # Run govulncheck
make sdk       # Regenerate server-side SDK from OpenAPI spec
make clean     # Remove build artifacts
```

Run a single test:
```bash
go test ./internal/kahuna/... -run TestWindowsUserDataTemplate -v
```

## Architecture

### Entry Point & Startup Sequence

`cmd/kahuna/main.go` → parses CLI args (config file path, debug flag, migrate flag) → initializes logger → creates `KahunaEngine` → runs DB preflight → either migrates schema or starts the HTTP server.

### Core Engine (`internal/kahuna/`)

All application logic lives in `internal/kahuna/`. Key files:

| File | Responsibility |
|------|---------------|
| `engine.go` | Orchestrates startup: registers all routers, initializes metrics, manages HTTP server lifecycle |
| `router.go` | Maps REST routes to handlers for 21+ resource types |
| `db.go` | MongoDB CRUD layer — singleton `GetDB()`, `Insert`, `Update`, `FindByID`, `FindByName`, `FindAll`, `Delete` |
| `cache.go` | FreeCache-backed in-memory cache, invalidated on writes |
| `config.go` | YAML config parsing, singleton with mutex |
| `http.go` | HTTP server with graceful shutdown (30min read, 60min write timeouts) |
| `migration.go` | Database schema evolution, runs at startup with `--migrate` flag |
| `metrics.go` | Prometheus metrics exporter (namespace: `kowabunga`) |
| `agents.go` | Remote agent connectivity tracking with keepalive monitoring |
| `cloud-init.go` | Generates cloud-init ISO images from YAML templates for Linux/Windows VMs |

### Resource Model

All resources embed the base `Resource` struct (`resource.go`) which provides `ID`, timestamps, schema version, name, and description. BSON `inline` tagging is used so derived types serialize correctly to MongoDB.

Generic factory functions (`FindResources[T]`, `FindResourceByID[T]`, `FindResourceByName[T]`) use Go type parameters to avoid repetitive query code across all 18+ resource types.

### Authentication

- JWT (signature + lifetime configurable)
- API Key (`X-API-Key` header)
- Bearer token (`Authorization: Bearer`)
- Role-based access: users have an allowed-routes list; login and password-reset endpoints bypass auth

### SDK Generation

The REST API surface is defined by an OpenAPI spec (v0.53.2). Run `make sdk` to regenerate the server-side SDK via `openapi-generator-cli`. The generated code lives alongside the hand-written handlers.

## Configuration

Default config path: `/etc/config/kahuna.yml` (production) or `./config/kahuna.yml` (local).

Key sections: `global.db` (MongoDB URI), `global.cache`, `global.jwt`, `global.smtp`, `global.bootstrap` (initial admin user + SSH public key), and `cloudinit` (paths to Linux/Windows template files).

Cloud-init templates must exist on disk at the paths specified in config before the server will start.

## Key External Dependencies

- `go.mongodb.org/mongo-driver/v2` — database
- `github.com/gorilla/mux` — HTTP routing
- `github.com/prometheus/client_golang` — metrics
- `github.com/kowabunga-cloud/common` — shared logging, protobuf, WebSocket utilities
- `libvirt.org/go/libvirtxml` — Libvirt XML for VM definitions
