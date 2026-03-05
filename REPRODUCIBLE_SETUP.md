# Reproducible Setup

This guide defines the canonical commands to verify that a fresh clone can be reproduced consistently.

## Scope

These checks validate:

- Go dependency and test reproducibility
- npm lockfile-based installs for root and package workspaces
- Docker Compose configuration resolution for production profiles

## Prerequisites

- Go (version defined by `go.mod`)
- Node.js 20+
- npm
- Docker Engine with Compose plugin
- GNU Make

## One-Command Validation

From repository root:

```bash
make smoke
```

`make smoke` runs:

- `go test -short ./...`
- `npm ci`
- `npm --prefix packages/core ci`
- `npm --prefix packages/privacy ci`
- `docker compose -f docker-compose.production.yml config >/dev/null`
- `docker compose -f docker-compose.1000nodes.yml config >/dev/null`

Success criteria:

- Command exits with status code `0`
- Final line includes `Smoke checks passed`

## CI Equivalence

GitHub Actions workflow `.github/workflows/reproducibility-check.yml` runs the same `make smoke` command after setting up Go and Node.

## Troubleshooting

- npm lockfile issues:
  - Ensure `package-lock.json`, `packages/core/package-lock.json`, and `packages/privacy/package-lock.json` are present and committed.
- Docker compose issues:
  - Verify Docker daemon is running and `docker compose version` succeeds.
- Go toolchain mismatch:
  - Use the Go version from `go.mod`.
