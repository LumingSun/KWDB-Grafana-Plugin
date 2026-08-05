import { expect, test } from '@grafana/plugin-e2e';

test('provisioned KWDB datasource passes health check', async ({ readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const response = await page.request.get(`/api/datasources/uid/${ds.uid}/health`);
  expect(response.ok()).toBeTruthy();
});

test('panel editor runs a KWDB query and returns a table frame', async ({ readProvisionedDataSource, panelEditPage }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  const page = panelEditPage.ctx.page;
  await page.getByRole('combobox', { name: 'Table' }).click();
  await page.getByRole('option', { name: 'charger_data (TIME SERIES TABLE)' }).click();
  await page.getByRole('combobox', { name: 'Time column' }).click();
  await page.getByRole('option', { name: 'ts' }).click();
  await page.getByRole('combobox', { name: 'Metric column 0' }).click();
  await page.getByRole('option', { name: 'voltage_v' }).click();

  const response = await panelEditPage.refreshPanel();
  expect(response.ok()).toBeTruthy();
});
