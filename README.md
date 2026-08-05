# KaiwuDB / KWDB Grafana data source plugin

This project is a Grafana data source plugin for [KaiwuDB](https://www.kaiwudb.com) and its open-source edition, KWDB. KWDB is the open-source version of KaiwuDB, and this plugin lets you query, visualize, and analyze KWDB time-series data directly in Grafana.

The plugin connects to KWDB through its PostgreSQL-compatible SQL service on port `26257`, executes read-only SQL queries, and converts the results into Grafana time series or table data frames.

## Plugin capabilities

- Visual SQL builder for downsampling, gapfill, latest values, window/event queries, plus a raw SQL mode.
- Time-series and table output formats with automatic KWDB type to Grafana Data Frame conversion.
- Metadata resources for listing tables and columns, with table type (`TIME SERIES TABLE` vs `TABLE`) surfaced in the query editor.
- Backend health check, Grafana Alerting, and annotation queries.
- Read-only query enforcement: only `SELECT`, `SHOW`, `EXPLAIN`, and `WITH` statements are allowed; multiple statements and write keywords are rejected.
- Configurable TLS root certificate for `verify-ca` / `verify-full` SSL modes.
- Result rows are capped at 100,000 per frame and a warning notice is attached when truncation happens.

> 本插件适用于 KaiwuDB 及其开源版本 KWDB，可以在 Grafana 中查询、展示和分析 KWDB 时序数据。

![KWDB TSDB query editor](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/main/src/img/screenshot-query-editor.png)

## Installation

- **From the Grafana plugin catalog:** install the plugin from the catalog once
  it is published, then create a data source of type `KWDB TSDB`.
- **From a ZIP:** build and sign the plugin, rename `dist` to the plugin ID,
  create a ZIP, and extract it into Grafana's plugin directory. See
  [Package a plugin](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin)
  for details.

## Configuration

The data source configuration page accepts the following settings:

| Field | Description |
| --- | --- |
| Host | KWDB host, for example `10.0.0.10`. |
| Port | KWDB PostgreSQL-compatible port, default `26257`. |
| Database | Database to query, default `defaultdb`. |
| User | Database user, default `root`. Prefer a read-only account. |
| SSL Mode | `disable`, `require`, `verify-ca`, or `verify-full`. |
| SSL Root Cert | Path to the CA certificate used with `verify-ca` / `verify-full`. |
| Password | Database password, stored in Grafana's secure JSON data. |

## Usage

The query editor supports a visual builder and a raw SQL editor:

- **Downsampling** uses `time_bucket` to aggregate metrics into time buckets.
- **Gapfill** adds `time_bucket_gapfill` with linear or constant interpolation.
- **Latest values** returns the latest row per tag set.
- **Window/Event** supports `TIME_WINDOW`, `SESSION_WINDOW`, `EVENT_WINDOW`,
  `COUNT_WINDOW`, and `STATE_WINDOW`.
- **Raw SQL** lets you write `SELECT`, `SHOW`, `EXPLAIN`, or `WITH` statements
  directly. The backend rejects multiple statements and write keywords.

Time-series output is converted to a wide Grafana frame automatically. Table
output returns the raw result columns.

Available query macros:

- `$__timeFilter(column)` expands to a timestamp range filter.
- `$__timeFrom` and `$__timeTo` expand to the dashboard time range.
- `$__timeGroup(column, 'interval')` expands to `time_bucket`.

## Related projects

- [KaiwuDB official website](https://www.kaiwudb.com)
- [KWDB source code on GitHub](https://github.com/KWDB/KWDB)
- [KWDB source code on Gitee](https://gitee.com/kwdb/kwdb)
- [KWDB documentation](https://www.kaiwudb.com/kaiwudb_docs/#/oss_dev/)
- [kwdb-tsdb-datasource repository](https://github.com/LumingSun/kwdb-tsdb-datasource)

## Getting started

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/grafana-plugin-sdk-for-go) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build plugin backend binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

### Frontend

1. Install dependencies

   ```bash
   npm install
   ```

2. Build plugin in development mode and run in watch mode

   ```bash
   npm run dev
   ```

3. Build plugin in production mode

   ```bash
   npm run build
   ```

4. Run the tests (using Jest)

   ```bash
   # Runs the tests and watches for changes, requires git init first
   npm run test

   # Exits after running all the tests
   npm run test:ci
   ```

5. Spin up a Grafana instance and run the plugin inside it (using Docker)

   ```bash
   npm run server
   ```

6. Run the E2E tests (using Playwright)

   ```bash
   # Spins up a Grafana instance first that we tests against
   npm run server

   # If you wish to start a certain Grafana version. The dev environment is tested on 13.1.0
   GRAFANA_VERSION=13.1.0 npm run server

   # Starts the tests
   npm run e2e
   ```

   E2E tests live in `tests-e2e/` and use `@grafana/plugin-e2e`. They expect
   Grafana on `http://localhost:3001` (or `GRAFANA_URL`) and use the system
   Chrome installation (`channel: 'chrome'`) instead of a downloaded Playwright
   browser.

7. Run the linter

   ```bash
   npm run lint

   # or

   npm run lint:fix
   ```

## Local Docker development environment

`docker-compose.yaml` starts KWDB (ports `26257` and `8080`) together with the
scaffolded Grafana dev container. Grafana is exposed on port `3001` by default
(override with `GRAFANA_PORT`), because the default `3000` is often already
occupied by other local services:

```bash
docker compose up -d
```

```bash
GRAFANA_PORT=3000 docker compose up -d
```

The compose stack disables Grafana 13's `dashboardNewLayouts` feature toggle
(`GF_FEATURE_TOGGLES_dashboardNewLayouts=false`) because the current
`@grafana/plugin-e2e` panel editor flow is not compatible with that layout.

On the first start, `scripts/init.sh` waits for KWDB to accept connections and
applies `scripts/init.sql`, which creates the `demo_ts` time-series database and
the `sensors` table and inserts recent sample rows. The initialization runs once
per persistent KWDB volume; a marker file inside the data volume prevents it from
re-running on later restarts.

Grafana auto-provisions a `KWDB TSDB` data source pointing at `host: kwdb`,
`database: demo_ts`, `user: root`, and `sslMode: disable`, so panels can query
the sample data immediately at <http://localhost:3001> (or the configured
`GRAFANA_PORT`).

The provisioned data source reads `KWDB_HOST`, `KWDB_PORT`, `KWDB_DATABASE`,
`KWDB_USER`, and `KWDB_PASSWORD` from the Grafana container environment. To run
Grafana against an existing KWDB instance instead of the bundled container:

```bash
KWDB_HOST=10.110.105.80 KWDB_PORT=26258 KWDB_DATABASE=ts_db \
  docker compose up --no-deps grafana
```

The compose stack builds from `grafana/grafana-enterprise:13.1.0` by default;
set `GRAFANA_IMAGE` / `GRAFANA_VERSION` to use a different Grafana image. When
developing on macOS, `mage build:backend` only produces Darwin binaries, so
also build the Linux backend binary that the Docker container loads:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
  -o dist/gpx_kwdb_tsdb_datasource_linux_arm64 \
  -tags arrow_json_stdlib -ldflags '-w -s' ./pkg
```

An optional backend integration test creates a temporary time-series table on a
real KWDB instance, exercises metadata, table/time-series frames, downsampling,
gapfill, and latest-value queries, then drops the table:

```bash
KWDB_TEST_DSN='postgresql://root:root@10.110.105.80:26258/defaultdb?sslmode=disable' \
  GOCACHE=/private/tmp/kwdb-gocache GOPATH=/private/tmp/kwdb-gopath \
  go test ./pkg/plugin -run TestRemoteKWDBIntegration -v -count=1
```

# Distributing your plugin

When distributing a Grafana plugin either within the community or privately the plugin must be signed so the Grafana application can verify its authenticity. This can be done with the `@grafana/sign-plugin` package.

_Note: It's not necessary to sign a plugin during development. The docker development environment that is scaffolded with `@grafana/create-plugin` caters for running the plugin without a signature._

## Initial steps

Before signing a plugin please read the Grafana [plugin publishing and signing criteria](https://grafana.com/legal/plugins/#plugin-publishing-and-signing-criteria) documentation carefully.

`@grafana/create-plugin` has added the necessary commands and workflows to make signing and distributing a plugin via the grafana plugins catalog as straightforward as possible.

Before signing a plugin for the first time please consult the Grafana [plugin signature levels](https://grafana.com/legal/plugins/#what-are-the-different-classifications-of-plugins) documentation to understand the differences between the types of signature level.

1. Create a [Grafana Cloud account](https://grafana.com/signup).
2. Make sure that the first part of the plugin ID matches the slug of your Grafana Cloud account.
   - _You can find the plugin ID in the `plugin.json` file inside your plugin directory. For example, if your account slug is `acmecorp`, you need to prefix the plugin ID with `acmecorp-`._
3. Create an Access Policy token with the `plugins:write` scope and save it as a
   GitHub Actions secret named `GRAFANA_ACCESS_POLICY_TOKEN`.
4. First submissions to the plugin catalog do not require a signature; Grafana
   assigns a signature level after review. Later versions are signed with
   `npm run sign`.

## Signing a plugin

### Using Github actions release workflow

If the plugin is using the github actions supplied with `@grafana/create-plugin` signing a plugin is included out of the box. The [release workflow](./.github/workflows/release.yml) builds, signs, and packages the plugin when a `v*` tag is pushed.

#### Push a version tag

To trigger the workflow we need to push a version tag to github. This can be achieved with the following steps:

1. Run `npm version <major|minor|patch>`
2. Run `git push origin master --follow-tags`

## Learn more

Below you can find source code for existing app plugins and other related documentation.

- [Basic data source plugin example](https://github.com/grafana/grafana-plugin-examples/tree/master/examples/datasource-basic#readme)
- [`plugin.json` documentation](https://grafana.com/developers/plugin-tools/reference/plugin-json)
- [How to sign a plugin?](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin)
