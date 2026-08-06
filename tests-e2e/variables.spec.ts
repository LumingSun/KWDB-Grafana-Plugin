import { expect, test } from '@grafana/plugin-e2e';

const e2eTable = process.env.KWDB_E2E_TABLE ?? 'demo_ts.sensors';
const e2eMetric = process.env.KWDB_E2E_METRIC ?? 'temperature';
const e2eVariableColumn = process.env.KWDB_E2E_VARIABLE_COLUMN ?? 'location';
const e2eVariableValues = (process.env.KWDB_E2E_VARIABLE_VALUES ?? 'room-a, room-b, room-c')
  .split(',')
  .map((value) => value.trim())
  .filter(Boolean);
const e2eVariableValue = e2eVariableValues[0];

test('Query variable returns KWDB tag values', async ({ readProvisionedDataSource, variableEditPage, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await variableEditPage.setVariableType('Query');
  await variableEditPage.datasource.set(ds.name);

  const editor = page.getByPlaceholder('Metric name or tags query');
  await editor.fill(`SELECT DISTINCT ${e2eVariableColumn} FROM ${e2eTable}`);
  await editor.blur();
  await variableEditPage.runQuery();

  await expect(variableEditPage).toDisplayPreviews(e2eVariableValues);
});

test('panel query expands template variable before backend', async ({
  readProvisionedDataSource,
  page,
  request,
  gotoDashboardPage,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const uid = `kwdb-e2e-variables-${Date.now()}`;
  const datasource = { type: ds.type, uid: ds.uid };
  const dashboard = {
    annotations: { list: [] },
    editable: true,
    id: null,
    panels: [
      {
        datasource,
        fieldConfig: { defaults: {}, overrides: [] },
        gridPos: { h: 8, w: 12, x: 0, y: 0 },
        id: 1,
        options: { showHeader: true },
        targets: [
          {
            datasource,
            format: 'table',
            mode: 'raw',
            rawSql: `SELECT ts, ${e2eMetric}, ${e2eVariableColumn} FROM ${e2eTable} WHERE ${e2eVariableColumn} = $location ORDER BY ts DESC LIMIT 10`,
            refId: 'A',
          },
        ],
        title: 'Variable panel',
        type: 'table',
      },
    ],
    schemaVersion: 39,
    tags: ['kwdb'],
    templating: {
      list: [
        {
          current: { text: e2eVariableValue, value: e2eVariableValue },
          datasource,
          definition: `SELECT DISTINCT ${e2eVariableColumn} FROM ${e2eTable}`,
          includeAll: false,
          label: 'Location',
          multi: false,
          name: 'location',
          options: e2eVariableValues.map((value) => ({ text: value, value })),
          query: `SELECT DISTINCT ${e2eVariableColumn} FROM ${e2eTable}`,
          refresh: 1,
          sort: 1,
          type: 'query',
        },
      ],
    },
    time: { from: 'now-30m', to: 'now' },
    title: 'KWDB Variables E2E',
    uid,
    version: 1,
  };

  const createResponse = await request.post('/api/dashboards/db', { data: { dashboard, overwrite: true } });
  expect(createResponse.ok()).toBeTruthy();

  const requestPromise = page.waitForRequest((requestItem) => {
    if (!requestItem.url().includes('/api/ds/query')) {
      return false;
    }
    const body = requestItem.postData() ?? '';
    return body.includes(`= '${e2eVariableValue}'`) && !body.includes('$location');
  });

  const dashboardPage = await gotoDashboardPage({ uid });
  await dashboardPage.refreshDashboard();

  await expect(requestPromise).toBeTruthy();
});
