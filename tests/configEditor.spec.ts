import { test, expect } from '@grafana/plugin-e2e';
import { IrisDataSourceOptions, IrisSecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByRole('textbox', { name: 'Host' })).toBeVisible();
  await expect(page.getByRole('textbox', { name: 'Namespace' })).toBeVisible();
  await expect(page.getByRole('textbox', { name: 'Password' })).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<IrisDataSourceOptions, IrisSecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Host' }).fill(ds.jsonData.host ?? 'iris');
  await page.getByRole('spinbutton', { name: 'Port' }).fill(String(ds.jsonData.port ?? 1972));
  await page.getByRole('textbox', { name: 'Namespace' }).fill(ds.jsonData.namespace ?? 'USER');
  await page.getByRole('textbox', { name: 'Username' }).fill(ds.jsonData.username ?? '_SYSTEM');
  await page.getByRole('textbox', { name: 'Password' }).fill(ds.secureJsonData?.password ?? '');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when password is missing', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<IrisDataSourceOptions, IrisSecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Host' }).fill(ds.jsonData.host ?? 'iris');
  await page.getByRole('spinbutton', { name: 'Port' }).fill(String(ds.jsonData.port ?? 1972));
  await page.getByRole('textbox', { name: 'Namespace' }).fill(ds.jsonData.namespace ?? 'USER');
  await page.getByRole('textbox', { name: 'Username' }).fill(ds.jsonData.username ?? '_SYSTEM');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'password is missing' });
});
