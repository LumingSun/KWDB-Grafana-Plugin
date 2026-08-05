import { expect, test } from '@grafana/plugin-e2e';

const e2eTable = process.env.KWDB_E2E_TABLE ?? 'demo_ts.sensors';
const e2eColumns = (process.env.KWDB_E2E_COLUMNS ?? 'ts, device_id, temperature')
  .split(',')
  .map((column) => column.trim())
  .filter(Boolean);

test('provisioned KWDB datasource passes health check', async ({ readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const response = await page.request.get(`/api/datasources/uid/${ds.uid}/health`);
  expect(response.ok()).toBeTruthy();
});

test('raw SQL query renders KWDB table data', async ({ readProvisionedDataSource, panelEditPage }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  const page = panelEditPage.ctx.page;
  await page.getByLabel('Mode').click();
  await page.getByRole('option', { name: 'Raw SQL' }).click();

  await page.getByLabel('Format').click();
  await page.getByRole('option', { name: 'Table' }).click();

  const editor = page.locator('.monaco-editor textarea').first();
  await editor.click();
  await page.keyboard.press('ControlOrMeta+A');
  await page.keyboard.type(`SELECT ${e2eColumns.join(', ')} FROM ${e2eTable} LIMIT 3`);

  const response = await panelEditPage.refreshPanel();
  expect(response.ok()).toBeTruthy();
  await expect(panelEditPage.panel.fieldNames).toContainText(e2eColumns);
});
