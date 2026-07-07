import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('textbox', { name: 'SQL' })).toBeVisible();
  await expect(panelEditPage.getQueryEditorRow('A').getByRole('radio', { name: 'Table' })).toBeVisible();
});

test('should trigger new query when row limit changes', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.getQueryEditorRow('A').getByRole('textbox', { name: 'SQL' }).fill('SELECT 1 AS value');
  const queryReq = panelEditPage.waitForQueryDataRequest();
  await panelEditPage.getQueryEditorRow('A').getByRole('spinbutton', { name: 'Row limit' }).fill('10');
  await panelEditPage.getQueryEditorRow('A').getByRole('spinbutton', { name: 'Row limit' }).blur();
  await expect(await queryReq).toBeTruthy();
});

test('data query should return a table from IRIS SQL', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.setVisualization('Table');

  const sqlEditor = panelEditPage.getQueryEditorRow('A').getByRole('textbox', { name: 'SQL' });
  await sqlEditor.fill('SELECT 1 AS value');
  const queryResponse = panelEditPage.waitForQueryDataResponse((response) => {
    return response.request().postData()?.includes('SELECT 1 AS value') ?? false;
  });
  await sqlEditor.blur();
  await expect(queryResponse).toBeOK();
  await expect(panelEditPage.panel.data).toContainText(['1']);
});
