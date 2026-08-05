import { expect, test } from '@grafana/plugin-e2e';

test('provisioned KWDB datasource passes health check', async ({ readProvisionedDataSource, gotoDataSourceConfigPage }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await gotoDataSourceConfigPage(ds.uid);

  const response = await configPage.saveAndTest();
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
  await page.keyboard.type('SELECT ts, charger_id, current_amp FROM ts_db.charger_data LIMIT 3');

  const response = await panelEditPage.refreshPanel();
  expect(response.ok()).toBeTruthy();
  await expect(panelEditPage.panel.fieldNames).toContainText(['ts', 'charger_id', 'current_amp']);
});
