# KWDB Grafana Data Source - Variable Ecosystem Implementation Plan

**Date:** 2026-08-06
**Design spec:** [2026-08-06-kwdb-variables-design.md](./2026-08-06-kwdb-variables-design.md)

## 1. Overview

This plan implements Phase 1 of the variable ecosystem for
`kwdb-tsdb-datasource`:

- Grafana Query variables backed by read-only KWDB SQL.
- Panel query interpolation through `applyTemplateVariables`.
- Unit and E2E coverage.

There are no Go backend changes.

## 2. Phase 1: Variable support frontend

### 2.1 Create `src/variableSupport.ts`

Add a new file with a `StandardVariableSupport` implementation:

```ts
import { StandardVariableQuery, StandardVariableSupport } from '@grafana/data';

import type { KwdbDataSource } from './datasource';
import type { KwdbQuery } from './types';

export class KwdbVariableSupport extends StandardVariableSupport<KwdbDataSource> {
  toDataQuery(query: StandardVariableQuery): KwdbQuery {
    return {
      refId: query.refId ?? 'variable-query',
      mode: 'raw',
      format: 'table',
      rawSql: query.query,
    };
  }
}
```

### 2.2 Update `src/datasource.ts`

- Import `getTemplateSrv` and `TemplateSrv` from `@grafana/runtime`.
- Import `KwdbVariableSupport`.
- Inject `TemplateSrv` in the constructor:

```ts
constructor(
  instanceSettings: DataSourceInstanceSettings<KwdbDataSourceOptions>,
  private readonly templateSrv: TemplateSrv = getTemplateSrv()
) {
  super(instanceSettings);
  this.variables = new KwdbVariableSupport();
}
```

- Implement `applyTemplateVariables`:

```ts
applyTemplateVariables(query: KwdbQuery, scopedVars: ScopedVars): KwdbQuery {
  if (!query.rawSql) {
    return query;
  }
  return {
    ...query,
    rawSql: this.templateSrv.replace(query.rawSql, scopedVars, 'sqlstring'),
  };
}
```

- Keep `getTables`, `getColumns`, `getDefaultQuery`, `annotations`, and
  `filterQuery` unchanged.

## 3. Phase 2: Unit tests

Add `tests/variableSupport.test.ts`:

- `KwdbVariableSupport.toDataQuery` maps `StandardVariableQuery` to a raw
  `KwdbQuery` with `format: 'table'`.
- `applyTemplateVariables` replaces a single-value variable through the
  injected `TemplateSrv`.
- `applyTemplateVariables` delegates to `TemplateSrv.replace` with the
  `sqlstring` format.
- `applyTemplateVariables` preserves `$__timeFilter`, `$__timeFrom`,
  `$__timeTo`, and `$__timeGroup`.
- `applyTemplateVariables` returns the query unchanged when `rawSql` is empty.

Mock `TemplateSrv.replace` in unit tests; do not depend on global Grafana
state.

## 4. Phase 3: E2E tests

### 4.1 Add `tests-e2e/variables.spec.ts`

Use the `variableEditPage` fixture:

- Set variable type to `Query`.
- Select the provisioned `KWDB TSDB` data source.
- Enter
  `SELECT DISTINCT {KWDB_E2E_VARIABLE_COLUMN} FROM {KWDB_E2E_TABLE}`, defaulting
  to `SELECT DISTINCT location FROM demo_ts.sensors`.
- Run the variable query and assert previews include the expected string values.

### 4.2 Add `provisioning/dashboards/kwdb-variables.json`

Provision a dashboard with a Query variable that uses the KWDB data source and
the default `location` query. The panel uses a Raw SQL query that references
`$location`, for example:

```sql
SELECT ts, temperature, device_id FROM demo_ts.sensors
WHERE location = $location
ORDER BY ts
```

Add an E2E test that opens the dashboard, spies on the query data request, and
asserts the request body contains the expanded `'room-a'` style value.

Register the new dashboard in
`provisioning/dashboards/dashboards.yml`.

## 5. Phase 4: Documentation

Update both `README.md` and `src/README.md`:

- Add a Variables section with Query variable examples.
- Document the default `sqlstring` formatting.
- Document `${var:raw}` and `${var:doublequote}` overrides.
- Document the string-field requirement for variable queries.
- Document that Ad Hoc filters and a custom variable editor are follow-up
  phases.

## 6. Phase 5: Verification

Run from the repository root with Node 22:

```bash
PATH=/Users/sunluming/.nvm/versions/node/v22.18.0/bin:$PATH yarn typecheck
PATH=/Users/sunluming/.nvm/versions/node/v22.18.0/bin:$PATH yarn lint
PATH=/Users/sunluming/.nvm/versions/node/v22.18.0/bin:$PATH yarn test:ci
PATH=/Users/sunluming/.nvm/versions/node/v22.18.0/bin:$PATH yarn build
```

E2E:

```bash
KWDB_E2E_TABLE=demo_ts.sensors \
KWDB_E2E_VARIABLE_COLUMN=location \
PATH=/Users/sunluming/.nvm/versions/node/v22.18.0/bin:$PATH yarn e2e
```

No Go verification is required because the backend is unchanged.

## 7. Acceptance criteria

- A Grafana Query variable can return options from a KWDB string column.
- A Raw SQL panel query can reference `$location` and receive expanded SQL
  string literals.
- Plugin time macros remain intact for backend expansion.
- Unit tests and E2E tests pass.
- Both READMEs document the new variable support.
