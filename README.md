# KaiwuDB / KWDB Grafana data source plugin

This plugin connects [Grafana](https://grafana.com) to KaiwuDB and its open-source edition, KWDB. It executes read-only SQL against the KWDB time-series engine and renders results as Grafana time series or table data.

![KWDB TSDB query editor](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/screenshot-query-editor.png)

> 本插件适用于 KaiwuDB 及其开源版本 KWDB，可以在 Grafana 中查询、展示和分析 KWDB 时序数据。

## Features

- Visual SQL builder for downsampling, gapfill, latest values, and window/event queries, plus a raw SQL editor.
- Grafana Query variables and template variable interpolation for dynamic dashboards.
- Time-series and table output formats with automatic KWDB type conversion.
- Table and column metadata browsing, including `TIME SERIES TABLE` vs `TABLE` detection.
- Backend health check, Grafana Alerting, and annotation queries.
- Read-only query enforcement: only `SELECT`, `SHOW`, `EXPLAIN`, and `WITH` statements are allowed.
- TLS root certificate support for `verify-ca` and `verify-full`.
- Result rows are capped at 100,000 per frame with a warning notice on truncation.

## Requirements

- Grafana 12.3 or newer
- KWDB or KaiwuDB with a PostgreSQL-compatible endpoint (default port `26257`)
- A database account with read access; prefer a dedicated read-only account

## Installation

- **From the Grafana plugin catalog:** install `lumingsun-kwdbtsdb-datasource` once it is published, then add a `KWDB TSDB` data source.
- **From a ZIP:** build and sign the plugin, rename `dist` to the plugin ID, create a ZIP, and extract it into Grafana's plugin directory. See the [packaging guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin).

## Configuration

| Field | Description |
| --- | --- |
| Host | KWDB host, for example `10.0.0.10`. |
| Port | PostgreSQL-compatible port, default `26257`. |
| Database | Database to query, default `defaultdb`. |
| User | Database user, default `root`. Prefer a read-only account. |
| SSL Mode | `disable`, `require`, `verify-ca`, or `verify-full`. |
| SSL Root Cert | Path to the CA certificate for `verify-ca` / `verify-full`. |
| Password | Database password, stored in Grafana secure JSON data. |

## Usage

The query editor supports five modes:

- **Downsampling** aggregates metrics with `time_bucket`.
- **Gapfill** fills time gaps with `time_bucket_gapfill` and interpolation.
- **Latest values** returns the latest row per tag set.
- **Window/Event** supports `TIME_WINDOW`, `SESSION_WINDOW`, `EVENT_WINDOW`, `COUNT_WINDOW`, and `STATE_WINDOW`.
- **Raw SQL** accepts read-only `SELECT`, `SHOW`, `EXPLAIN`, or `WITH` statements.

Available macros:

- `$__timeFilter(column)` expands to a timestamp range filter.
- `$__timeFrom` and `$__timeTo` expand to the dashboard time range.
- `$__timeGroup(column, 'interval')` expands to `time_bucket`.

## Variables

The data source supports Grafana Query variables and template variable
interpolation. Create a Query variable with read-only SQL that returns a string
column, for example:

```sql
SELECT DISTINCT location FROM demo_ts.sensors
```

Use the variable in a Raw SQL query:

```sql
SELECT ts, temperature, location FROM demo_ts.sensors
WHERE location = $location
```

Variable values are formatted as SQL string literals by default. Use
`${var:raw}` for raw values and `${var:doublequote}` for double-quoted
identifiers. Variable queries must return at least one string field; numeric-
or timestamp-only results are not displayed as options. In the current
version, interpolation applies to the final `rawSql`, so use Raw SQL mode when
writing variables.

## Development

See [DEVELOPMENT.md](https://github.com/LumingSun/kwdb-tsdb-datasource/blob/master/DEVELOPMENT.md)
for local development, testing, and packaging instructions.

## Publishing

- Plugin ID: `lumingsun-kwdbtsdb-datasource`; it must match the Grafana Cloud organization slug `lumingsun`.
- First submissions do not require a signature. After review, versions are signed and released through the `Release` GitHub Actions workflow.
- Add `GRAFANA_ACCESS_POLICY_TOKEN` (Access Policy token with `plugins:write`) to the repository secrets.
- Run the [plugin validator](https://github.com/grafana/plugin-validator) before submitting.

Full instructions are in the [Grafana plugin publishing guide](https://grafana.com/developers/plugin-tools/publish-a-plugin).

## License

Apache License 2.0. See
[LICENSE](https://github.com/LumingSun/kwdb-tsdb-datasource/blob/master/LICENSE).
