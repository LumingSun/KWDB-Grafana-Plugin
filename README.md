# KaiwuDB / KWDB Grafana Data Source Plugin

This plugin connects [Grafana](https://grafana.com) to KaiwuDB and its
open-source edition, KWDB. It executes read-only SQL against the KWDB
time-series engine through the PostgreSQL wire protocol and renders results as
Grafana time series or table data.

![KWDB TSDB query editor](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/screenshot-query-editor.png)

> 本插件适用于 KaiwuDB 及其开源版本 KWDB，可以在 Grafana 中查询、展示和分析
> KWDB 时序数据。

## Features

- Visual SQL builder for downsampling, gapfill, latest values, and window/event
  queries, plus a raw SQL editor.
- Grafana Query variables and template variable interpolation.
- Time-series and table output formats with automatic KWDB type conversion.
- Table and column metadata browsing, including `TIME SERIES TABLE` vs `TABLE`
  detection.
- Backend health check, Grafana Alerting, and annotations.
- Read-only query enforcement.
- TLS root certificate support.
- 100,000-row result cap with a warning notice.

## Quick Start

1. Install the plugin and add a `KWDB TSDB` data source.
2. Configure the KWDB host, port, database, user, and password.
3. Click **Save & test**.
4. Create a dashboard and add a panel using the KWDB data source.

See the [User Guide](docs/USER_GUIDE.md) for detailed installation,
configuration, query editor, variable, alerting, and troubleshooting
instructions.

## Requirements

- Grafana 12.3 or newer
- KWDB or KaiwuDB with a PostgreSQL-compatible endpoint (default port `26257`)
- A database account with read access; prefer a dedicated read-only account

## Installation

- **From the Grafana plugin catalog:** install
  `lumingsun-kwdbtsdb-datasource` once it is published, then add a `KWDB TSDB`
  data source.
- **From a ZIP:** build and sign the plugin, rename `dist` to the plugin ID,
  create a ZIP, and extract it into Grafana's plugin directory.
- **Local development:** `docker compose up -d` starts a provisioned KWDB and
  Grafana environment with a demo dashboard.

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

## Query Editor

The query editor supports five modes:

- **Downsampling** aggregates metrics with `time_bucket`.
- **Gapfill** fills time gaps with `time_bucket_gapfill` and interpolation.
- **Latest values** returns the latest row per tag set.
- **Window/Event** supports `TIME_WINDOW`, `SESSION_WINDOW`, `EVENT_WINDOW`,
  `COUNT_WINDOW`, and `STATE_WINDOW`.
- **Raw SQL** accepts read-only `SELECT`, `SHOW`, `EXPLAIN`, or `WITH`
  statements.

The generated SQL is shown in a full-width, vertically resizable preview.

## Macros

- `$__timeFilter(column)` expands to a timestamp range filter.
- `$__timeFrom` and `$__timeTo` expand to the dashboard time range.
- `$__timeGroup(column, 'interval')` expands to `time_bucket`.

## Variables

Create a Grafana Query variable with read-only SQL that returns a string
column:

```sql
SELECT DISTINCT station_code FROM charger_data
```

Use the variable in a Raw SQL query:

```sql
SELECT ts, voltage_v, station_code
FROM charger_data
WHERE station_code = $station
```

Variable values are formatted as SQL string literals by default. Use
`${var:raw}` for raw values and `${var:doublequote}` for double-quoted
identifiers. Variable queries must return at least one string field.

## Alerting and Annotations

The plugin supports Grafana Alerting and annotations through the same backend
query pipeline.

## Development and Publishing

- See [DEVELOPMENT.md](DEVELOPMENT.md) for local development, testing, and
  packaging.
- Plugin ID: `lumingsun-kwdbtsdb-datasource`
- Full publishing instructions:
  [Grafana plugin publishing guide](https://grafana.com/developers/plugin-tools/publish-a-plugin)

## License

Apache License 2.0. See [LICENSE](LICENSE).
