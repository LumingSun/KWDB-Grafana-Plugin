# Changelog

## 1.0.0 (Unreleased)

**Implemented enhancements:**

- Visual SQL builder for downsampling, gapfill, latest values, window/event queries, and raw SQL.
- Time-series and table output formats with automatic KWDB type to Grafana Data Frame conversion.
- Metadata resources for listing tables and columns, including time-series table type detection.
- Backend health check, Grafana Alerting, and annotation query support.
- Read-only SQL enforcement with multi-statement and write-keyword rejection.
- Microsecond-precision time macros, connection pooling, concurrent query execution, and a 100,000-row result cap.
- TLS root certificate configuration for `verify-ca` and `verify-full` SSL modes.
- Provisioned KWDB demo environment with sample dashboard and Playwright E2E smoke tests.

**Breaking changes:**

- None.
