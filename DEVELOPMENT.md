# Development

This file contains development, testing, and packaging instructions for the
KWDB TSDB Grafana data source plugin.

## Prerequisites

- Node.js 22 (the repository pins it in `.nvmrc`)
- Go 1.26 or newer
- Docker and Docker Compose
- Yarn 1.22.22 via Corepack

## Install, build, and test

```bash
yarn install --frozen-lockfile
yarn typecheck
yarn lint
yarn test:ci
yarn build
```

Backend verification:

```bash
go test ./pkg/...
go vet ./pkg/...
golangci-lint run ./pkg/...
```

## Local Grafana environment

```bash
docker compose up -d
```

Grafana is exposed on `http://localhost:3001` (override with `GRAFANA_PORT`).
The compose stack starts KWDB, seeds the `demo_ts.sensors` table, provisions the
`KWDB TSDB` data source, and loads the `KWDB TSDB Sensors` demo dashboard.

To point Grafana at an existing KWDB instance:

```bash
KWDB_HOST=10.110.105.80 KWDB_PORT=26258 KWDB_DATABASE=ts_db \
  docker compose up -d --no-deps grafana
```

## End-to-end tests

```bash
KWDB_E2E_TABLE=demo_ts.sensors yarn e2e
```

E2E tests live in `tests-e2e/` and default to the provisioned
`demo_ts.sensors` table. Override the table, columns, and metric for another
environment:

```bash
KWDB_E2E_TABLE=ts_db.charger_data \
KWDB_E2E_COLUMNS='ts, charger_id, current_amp, voltage_v' \
KWDB_E2E_METRIC=voltage_v \
yarn e2e
```

The local Playwright config uses the system Chrome installation and expects
Grafana on `http://localhost:3001`.

## Cross-platform backend binaries

`mage -v build:backend` only produces Darwin binaries. Build the Linux binary
that the Docker container loads:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
  -o dist/gpx_kwdb_tsdb_datasource_linux_arm64 \
  -tags arrow_json_stdlib -ldflags '-w -s' ./pkg
```

The GitHub release workflow builds and packages all platform binaries
automatically.
