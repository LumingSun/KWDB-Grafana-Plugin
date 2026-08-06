# KWDB TSDB Grafana Data Source User Guide

This guide explains how to install and use the KWDB TSDB Grafana data source
plugin. It covers data source configuration, the visual query editor, SQL
macros, template variables, alerting, annotations, and output formats.

## 1. Overview

The KWDB TSDB data source connects Grafana to KaiwuDB and its open-source
edition KWDB through the PostgreSQL wire protocol (default port `26257`). It
executes read-only SQL against KWDB time-series tables and renders results as
Grafana time series or table data.

Key capabilities:

- Visual SQL builder for downsampling, gapfill, latest values, and window/event
  queries, plus a raw SQL editor.
- Grafana Query variables and template variable interpolation.
- Table and column metadata discovery, including time-series table and tag
  detection.
- Backend health check, Grafana Alerting, and annotations.
- Read-only SQL enforcement.
- TLS root certificate support.

> 本插件适用于 KaiwuDB 及其开源版本 KWDB，可以在 Grafana 中查询、展示和分析
> KWDB 时序数据。

## 2. Prerequisites

- Grafana 12.3 or newer.
- KWDB or KaiwuDB with a PostgreSQL-compatible endpoint.
- A database account with read access. Prefer a dedicated read-only account.

| Component | Requirement |
| --- | --- |
| Grafana | 12.3 or newer |
| KWDB / KaiwuDB | PostgreSQL-compatible endpoint, default port `26257` |
| Database account | Read access; a dedicated read-only account is recommended |

## 3. Installation

### 3.1 Install from the Grafana plugin catalog

Once the plugin is published, install `lumingsun-kwdbtsdb-datasource` from the
Grafana plugin catalog, then add a `KWDB TSDB` data source.

> TODO: Add screenshot: Grafana plugin catalog showing KWDB TSDB.

### 3.2 Install with grafana-cli

When the plugin is published, you can install it from the command line:

```bash
grafana-cli plugins install lumingsun-kwdbtsdb-datasource
```

Restart Grafana after installation.

### 3.3 Manual ZIP installation

1. Build and sign the plugin.
2. Rename `dist` to `lumingsun-kwdbtsdb-datasource`.
3. Create a ZIP archive.
4. Extract it into Grafana's plugin directory.
5. Restart Grafana.

See the
[Grafana plugin packaging guide](https://grafana.com/developers/plugin-tools/publish-a-plugin/package-a-plugin).

### 3.4 Local development environment

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

## 4. Add a data source

1. Open Grafana and navigate to **Connections** > **Data sources**.
2. Click **Add data source**.
3. Search for **KWDB TSDB**.
4. Fill in the connection settings.
5. Click **Save & test**.

> TODO: Add screenshot: KWDB TSDB data source configuration page.

### 4.1 Configuration fields

| Field | Description |
| --- | --- |
| Host | KWDB host, for example `10.0.0.10`. |
| Port | PostgreSQL-compatible port, default `26257`. |
| Database | Database to query, default `defaultdb`. |
| User | Database user, default `root`. Prefer a read-only account. |
| SSL Mode | `disable`, `require`, `verify-ca`, or `verify-full`. |
| SSL Root Cert | Path to the CA certificate for `verify-ca` / `verify-full`. |
| Password | Database password, stored in Grafana secure JSON data. |

### 4.2 TLS/SSL

For `verify-ca` and `verify-full`, provide the CA certificate path in
**SSL Root Cert**. The certificate path is passed to the PostgreSQL connection
as `sslrootcert`.

### 4.3 Save and test

Click **Save & test** to verify connectivity. The backend executes `SELECT 1`
and reports whether the data source is working.

> TODO: Add screenshot: Save & test success message.

## 5. Build a dashboard

1. Create a new dashboard.
2. Add a panel.
3. Select the **KWDB TSDB** data source.
4. Choose a query mode in the query editor.
5. Refresh the panel and verify the data.

> TODO: Add screenshot: example KWDB dashboard.

## 6. Query editor

The query editor provides five modes:

- **Downsampling**
- **Gapfill**
- **Latest Values**
- **Window/Event**
- **Raw SQL**

Each visual mode generates SQL automatically and shows a read-only SQL preview.
The SQL preview fills the available width and can be resized vertically with
the drag handle below the editor.

### 6.1 Downsampling

Downsampling aggregates metrics into fixed time buckets with `time_bucket`.
Select a table, time column, interval, tag columns for `GROUP BY`, and one or
more metrics.

Aggregations: `avg`, `sum`, `count`, `min`, `max`, `stddev`, `none`.

Example generated SQL:

```sql
SELECT time_bucket("ts", '5m') AS "time",
       "station_code",
       avg("voltage_v") AS "avg_voltage_v"
FROM "charger_data"
WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo
GROUP BY "time", "station_code"
ORDER BY "time", "station_code"
```

> TODO: Add screenshot: downsampling query editor.

### 6.2 Gapfill

Gapfill fills missing time buckets with `time_bucket_gapfill` and
`interpolate`. Supported interpolation modes:

- `linear`
- `PREV`
- `NEXT`
- `constant`
- `NULL`

Example generated SQL:

```sql
SELECT time_bucket_gapfill("ts", '5m') AS "time",
       interpolate(avg("voltage_v"), 'linear') AS "avg_voltage_v"
FROM "charger_data"
WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo
GROUP BY "time"
ORDER BY "time"
```

> TODO: Add screenshot: gapfill query editor.

### 6.3 Latest values

Latest values returns the latest row per tag set. Supported functions:

- `last`
- `last_row`
- `first`
- `first_row`

Example generated SQL:

```sql
SELECT "station_code",
       last("voltage_v") AS "latest_voltage_v",
       last("ts") AS "time"
FROM "charger_data"
WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo
GROUP BY "station_code"
```

> TODO: Add screenshot: latest values query editor.

### 6.4 Window/Event

Window/Event queries use KWDB window functions:

- `TIME_WINDOW`
- `SESSION_WINDOW`
- `EVENT_WINDOW`
- `COUNT_WINDOW`
- `STATE_WINDOW`

Example generated SQL:

```sql
SELECT TIME_WINDOW("ts", '1h', '15m') AS "time",
       avg("voltage_v") AS "avg_voltage_v"
FROM "charger_data"
WHERE "ts" >= $__timeFrom AND "ts" <= $__timeTo
GROUP BY "time"
ORDER BY "time"
```

> TODO: Add screenshot: window/event query editor.

### 6.5 Raw SQL

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

## 7. Time macros

| Macro | Expansion |
| --- | --- |
| `$__timeFilter(column)` | `column >= '...'::TIMESTAMP AND column <= '...'::TIMESTAMP` |
| `$__timeFrom` | Dashboard time range start as a timestamp literal |
| `$__timeTo` | Dashboard time range end as a timestamp literal |
| `$__timeGroup(column, 'interval')` | `time_bucket(column, 'interval')` |

Subsecond precision is preserved in the expanded timestamp literals.

## 8. Variables

The data source supports Grafana Query variables and template variable
interpolation.

### 8.1 Create a Query variable

Create a Query variable with read-only SQL that returns a string column:

```sql
SELECT DISTINCT station_code FROM charger_data
```

The first string column in the result becomes the variable options.

> TODO: Add screenshot: variable editor with query preview.

### 8.2 Use a variable in a query

Variables are interpolated into the final `rawSql`. In the current version,
use Raw SQL mode when writing variables:

```sql
SELECT ts, voltage_v, station_code
FROM charger_data
WHERE station_code = $station
```

### 8.3 Variable formatting

Variable values are formatted as SQL string literals by default:

- `$station` becomes `'station_A'`.
- Multi-value `$station` becomes `'station_A','station_B'`.

Override formatting inline:

- `${var:raw}` for raw values.
- `${var:doublequote}` for double-quoted identifiers.

### 8.4 Limitations

- Variable queries must return at least one string field. Numeric- or
  timestamp-only results are not displayed as options.
- Ad Hoc filter variables are not supported yet.
- The visual builder controls are not variable-aware; use Raw SQL mode when
  writing variables.

## 9. Alerting and annotations

The plugin declares support for Grafana Alerting and annotations. Alerting and
annotation queries use the same backend query pipeline and read-only SQL
enforcement.

## 10. Output formats

Each query returns one of two formats:

- **Time series**: the backend finds the time column, sorts by time, converts
  tag columns into labels, and produces a wide time-series frame.
- **Table**: the result is returned as a long table frame.

Result rows are capped at 100,000 rows per frame. When the cap is reached, the
frame includes a warning notice.

## 11. Read-only safety

The backend only allows SQL starting with `SELECT`, `SHOW`, `EXPLAIN`, or
`WITH`. It rejects:

- Multi-statement SQL.
- `SELECT INTO`.
- DML inside `WITH`.
- `CREATE`, `DROP`, `ALTER`, `INSERT`, `UPDATE`, `DELETE`, `TRUNCATE`, and
  other write keywords.

Keywords inside string literals, quoted identifiers, and comments are ignored.

## 12. Troubleshooting

### Health check fails

- Verify the KWDB host, port, database, user, and password.
- Confirm the port is reachable from the Grafana backend.
- Check KWDB account permissions.

### Table list is empty

- Confirm the configured database exists.
- Confirm the account can run `SHOW TABLES`.

### Variable preview is empty

- Confirm the variable query is read-only.
- Confirm the query returns at least one string column.

### Results are truncated

- The backend caps results at 100,000 rows per frame.
- Add `LIMIT` or narrow the time range.
