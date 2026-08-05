# Provisioned test environment

This directory contains the provisioning files that let reviewers and
contributors run the plugin end to end without manual setup:

- `datasources/datasources.yml` registers the `KWDB TSDB` data source. The
  connection settings come from `KWDB_HOST`, `KWDB_PORT`, `KWDB_DATABASE`,
  `KWDB_USER`, and `KWDB_PASSWORD`.
- `dashboards/dashboards.yml` loads the sample dashboard.
- `dashboards/kwdb-sensors.json` is a demo dashboard with a time series panel
  and a table panel that query `demo_ts.sensors`.

The `docker-compose.yaml` at the repository root starts a KWDB container, seeds
`demo_ts.sensors` through `scripts/init.sql`, and provisions Grafana. After
`docker compose up -d`, open <http://localhost:3001> and the dashboard appears
under the `KWDB` folder.

For more information see [Provision dashboards and data sources](https://grafana.com/tutorials/provision-dashboards-and-data-sources/)
