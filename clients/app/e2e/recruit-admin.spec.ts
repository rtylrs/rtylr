import { test, expect, type Page } from './fixtures';
import { stubPortalSession, PORTAL_URL } from './portal-session';

const career = {
  id: 42,
  ownerId: 1,
  slug: 'backend-engineer',
  title: 'Backend Engineer',
  department: 'Engineering',
  location: 'Remote',
  employmentType: 'Full-time',
  workMode: 'Remote',
  seniority: 'Senior',
  details: 'Build things.',
  requirements: '',
  salaryMin: 0,
  salaryMax: 0,
  salaryCurrency: 'USD',
  isPublished: true,
  isActive: true,
  createdAt: '2026-04-01T00:00:00Z',
  updatedAt: '2026-04-01T00:00:00Z',
};
async function APIs(page: Page) {
  let careers = [career];
  let status = 'new';
  await page.route('http://localhost:9999/recruit/customization', (r) =>
    r.fulfill({ json: { customization: null } })
  );
  await page.route('http://localhost:9999/recruit/careers', async (r) => {
    const method = r.request().method();
    if (method === 'POST') {
      careers = [...careers, { ...career, ...r.request().postDataJSON(), id: 43 }];
      return r.fulfill({ json: { career: careers[1] } });
    }
    return r.fulfill({ json: { careers } });
  });
  await page.route('http://localhost:9999/recruit/submissions?*', (r) =>
    r.fulfill({
      json: {
        submissions: [
          {
            id: 101,
            careerId: 42,
            candidateName: 'Jamie Test',
            email: 'jamie@example.test',
            status,
            createdAt: '2026-04-15T12:00:00Z',
          },
        ],
      },
    })
  );
  await page.route('http://localhost:9999/recruit/submissions/status', (r) => {
    status = r.request().postDataJSON().status;
    return r.fulfill({ json: { message: 'updated' } });
  });
}

test('owner creates a role and list updates without leaving Portal', async ({ page, context }) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['recruit'] });
  await APIs(page);
  await page.goto(`${PORTAL_URL}/recruit/careers`);
  await expect(page.getByText('Backend Engineer', { exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'New role', exact: true }).click();
  const dialog = page.getByRole('dialog', { name: 'New role' });
  await dialog.getByRole('textbox', { name: 'Title *', exact: true }).fill('Migration Specialist');
  await dialog.getByLabel('Role description').fill('Manage migrations.');
  const mutation = page.waitForRequest(
    (r) => r.url().endsWith('/recruit/careers') && r.method() === 'POST'
  );
  await dialog.getByRole('button', { name: 'Create role' }).click();
  expect((await mutation).postDataJSON()).toMatchObject({
    title: 'Migration Specialist',
    details: 'Manage migrations.',
  });
  await expect(dialog).not.toBeVisible();
  await expect(page.getByText('Migration Specialist', { exact: true })).toBeVisible();
  await expect(page).toHaveURL(`${PORTAL_URL}/recruit/careers`);
});

test('candidate status persists through refetch', async ({ page, context }) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['recruit'] });
  await APIs(page);
  await page.goto(`${PORTAL_URL}/recruit/submissions`);
  await expect(page.getByText('Jamie Test')).toBeVisible();
  const mutation = page.waitForRequest(
    (r) => r.url().endsWith('/recruit/submissions/status') && r.method() === 'PUT'
  );
  await page.getByRole('combobox', { name: 'Status', exact: true }).selectOption('reviewed');
  expect((await mutation).postDataJSON()).toEqual({ id: 101, status: 'reviewed' });
  await expect(page.getByRole('combobox', { name: 'Status', exact: true })).toHaveValue('reviewed');
  await page.reload();
  await expect(page.getByRole('combobox', { name: 'Status', exact: true })).toHaveValue('reviewed');
});

test('jobs viewer sees roles without creation or candidate requests', async ({ page, context }) => {
  await stubPortalSession(page, context, {
    isOwner: false,
    appIds: ['recruit'],
    permissions: ['recruit:jobs:view'],
  });
  await APIs(page);
  let candidateRequests = 0;
  page.on('request', (r) => {
    if (r.url().includes('/recruit/submissions')) candidateRequests++;
  });
  await page.goto(`${PORTAL_URL}/recruit/careers`);
  await expect(page.getByText('Backend Engineer', { exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'New role', exact: true })).toHaveCount(0);
  expect(candidateRequests).toBe(0);
});

test('failed career request has a working retry', async ({ page, context }) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['recruit'] });
  await APIs(page);
  await page.route('http://localhost:9999/recruit/careers', (r) =>
    r.fulfill({ status: 500, json: { message: 'Unavailable' } })
  );
  await page.goto(`${PORTAL_URL}/recruit/careers`);
  const retry = page.getByRole('button', { name: 'Retry' });
  await expect(retry).toBeVisible({ timeout: 15000 });
  await page.unroute('http://localhost:9999/recruit/careers');
  await page.route('http://localhost:9999/recruit/careers', (r) =>
    r.fulfill({ json: { careers: [career] } })
  );
  await retry.click();
  await expect(page.getByText('Backend Engineer', { exact: true })).toBeVisible();
});
