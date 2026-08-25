# KaiwuDB / KWDB Grafana Data Source Plugin

This plugin connects [Grafana](https://grafana.com) to KaiwuDB and its
open-source edition, KWDB. It executes read-only SQL against the KWDB
time-series engine through the PostgreSQL wire protocol and renders results as
Grafana time series or table data.

> 本插件适用于 KaiwuDB 及其开源版本 KWDB，可以在 Grafana 中查询、展示和分析
> KWDB 时序数据。
![example](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/example.png)
## Features

- Visual SQL builder for downsampling, gapfill, latest values, and window/event
  queries, plus a raw SQL editor.
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
5. Choose a query mode and refresh the panel.

## Requirements

- Grafana 12.3 or newer
- KWDB or KaiwuDB with a PostgreSQL-compatible endpoint (default port `26257`)
- A database account with read access; prefer a dedicated read-only account

| Component | Requirement |
| --- | --- |
| Grafana | 12.3 or newer |
| KWDB / KaiwuDB | PostgreSQL-compatible endpoint, default port `26257` |
| Database account | Read access; a dedicated read-only account is recommended |

## Installation

### Grafana plugin catalog

Install `lumingsun-kwdbtsdb-datasource` from the Grafana plugin catalog once it
is published, then add a `KWDB TSDB` data source.

### grafana-cli

When the plugin is published, install it from the command line:

```bash
grafana-cli plugins install lumingsun-kwdbtsdb-datasource
```

Restart Grafana after installation.

### Manual ZIP installation

1. Build and sign the plugin.
2. Rename `dist` to `lumingsun-kwdbtsdb-datasource`.
3. Create a ZIP archive.
4. Extract it into Grafana's plugin directory.
5. Restart Grafana.

See the
[Grafana plugin packaging guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin).

### Local development environment

The repository includes a Docker Compose stack that starts KWDB, seeds the
`demo_ts.sensors` table, provisions the `KWDB TSDB` data source, and loads a
demo dashboard:

```bash
docker compose up -d
```

Grafana is available at `http://localhost:3001` by default. Override the port
with `GRAFANA_PORT`.

To point the provisioned data source at an existing KWDB instance:

```bash
KWDB_HOST=10.110.105.80 KWDB_PORT=26258 KWDB_DATABASE=ts_db \
  docker compose up -d --no-deps grafana
```

## Configuration

1. Open Grafana and navigate to **Connections** > **Data sources**.
2. Click **Add data source**.
3. Search for **KWDB TSDB**.
4. Fill in the connection settings.
5. Click **Save & test**.

![KWDB Datasource](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/kwdb-datasource.png)

| Field | Description |
| --- | --- |
| Host | KWDB host, for example `10.0.0.10`. |
| Port | PostgreSQL-compatible port, default `26257`. |
| Database | Database to query, default `defaultdb`. |
| User | Database user, default `root`. Prefer a read-only account. |
| SSL Mode | `disable`, `require`, `verify-ca`, or `verify-full`. |
| SSL Root Cert | Path to the CA certificate for `verify-ca` / `verify-full`. |
| Password | Database password, stored in Grafana secure JSON data. |

### TLS/SSL

For `verify-ca` and `verify-full`, provide the CA certificate path in
**SSL Root Cert**. The certificate path is passed to the PostgreSQL connection
as `sslrootcert`.

### Save and test

Click **Save & test** to verify connectivity. The backend executes `SELECT 1`
and reports whether the data source is working.

![test-datasource](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/save-test.png)

## Build a dashboard

1. Create a new dashboard.
2. Add a panel.
3. Select the **KWDB TSDB** data source.
4. Choose a query mode in the query editor.
5. Refresh the panel and verify the data.

## Query editor

The query editor provides five modes:

- **Downsampling**
- **Gapfill**
- **Latest Values**
- **Window/Event**
- **Raw SQL**

Each visual mode generates SQL automatically and shows a read-only SQL preview.
The SQL preview fills the available width and can be resized vertically with
the drag handle below the editor.

### Downsampling

Downsampling aggregates metrics into fixed time buckets with `time_bucket`.
Select a table, time column, interval, tag columns for `GROUP BY`, and one or
more metrics.

Aggregations: `avg`, `sum`, `count`, `min`, `max`, `stddev`, `none`.

![downsampling example](https://github.com/LumingSun/kwdb-tsdb-datasource/raw/master/src/img/downsampling-example.png)

### Gapfill

Gapfill fills missing time buckets with `time_bucket_gapfill` and
`interpolate`. Supported interpolation modes:

- `linear`
- `PREV`
- `NEXT`
- `constant`
- `NULL`


### Latest values

Latest values returns the latest row per tag set. Supported functions:

- `last`
- `last_row`
- `first`
- `first_row`

Results are split into one series per tag combination, so Stat and Gauge
panels display one value per device without extra panel configuration:

- With the **Time series** format the backend converts tag columns into
  labels on a single wide frame.
- With the **Table** format the backend returns one frame per tag value,
  keeping every original column so tables still render normally.

Disable the **Split per tag** toggle in the Latest Values editor to return
a single merged table frame instead. As a panel-side alternative, Gauge/Stat
panels can also set **Value options → Show: All values** to expand grouped
rows into one value per device.


### Window/Event

Window/Event queries use KWDB window functions:

- `TIME_WINDOW`
- `SESSION_WINDOW`
- `EVENT_WINDOW`
- `COUNT_WINDOW`
- `STATE_WINDOW`


### Raw SQL

Raw SQL accepts read-only statements that start with `SELECT`, `SHOW`,
`EXPLAIN`, or `WITH`. Multi-statement queries and write statements are
rejected.

Example:

```sql
SELECT ts, voltage_v, station_code
FROM charger_data
WHERE $__timeFilter(ts)
ORDER BY ts
LIMIT 100
```

## Time macros

| Macro | Expansion |
| --- | --- |
| `$__timeFilter(column)` | `column >= '...'::TIMESTAMP AND column <= '...'::TIMESTAMP` |
| `$__timeFrom` | Dashboard time range start as a timestamp literal |
| `$__timeTo` | Dashboard time range end as a timestamp literal |
| `$__timeGroup(column, 'interval')` | `time_bucket(column, 'interval')` |

Subsecond precision is preserved in the expanded timestamp literals.

## Alerting and annotations

The plugin supports Grafana Alerting and annotations. Alerting and annotation
queries use the same backend query pipeline and read-only SQL enforcement.

## Output formats

Each query returns one of two formats:

- **Time series**: the backend finds the time column, sorts by time, converts
  tag columns into labels, and produces a wide time-series frame.
- **Table**: the result is returned as a long table frame. Latest-values
  queries with tag columns return one frame per tag combination by default,
  so every device becomes its own series for Gauge/Stat panels. Disable
  the **Split per tag** toggle to return a single merged table frame.

Result rows are capped at 100,000 rows per frame. When the cap is reached, the
frame includes a warning notice.

## Read-only safety

The backend only allows SQL starting with `SELECT`, `SHOW`, `EXPLAIN`, or
`WITH`. It rejects:

- Multi-statement SQL.
- `SELECT INTO`.
- DML inside `WITH`.
- `CREATE`, `DROP`, `ALTER`, `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE`, and
  other write keywords.

Keywords inside string literals, quoted identifiers, and comments are ignored.

## Troubleshooting

### Health check fails

- Verify the KWDB host, port, database, user, and password.
- Confirm the port is reachable from the Grafana backend.
- Check KWDB account permissions.

### Table list is empty

- Confirm the configured database exists.
- Confirm the account can run `SHOW TABLES`.

### Results are truncated

- The backend caps results at 100,000 rows per frame.
- Add `LIMIT` or narrow the time range.

## Development and publishing

- See [DEVELOPMENT.md](DEVELOPMENT.md) for local development, testing, and
  packaging.
- Plugin ID: `lumingsun-kwdbtsdb-datasource`
- First submissions do not require a signature. After review, versions are
  signed and released through the `Release` GitHub Actions workflow.
- Full publishing instructions:
  [Grafana plugin publishing guide](https://grafana.com/developers/plugin-tools/publish-a-plugin)

## License

Apache License 2.0. See [LICENSE](LICENSE).
