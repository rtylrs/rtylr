import {
  encodeGetDealBoardResponse,
  encodeGetDealStatsResponse,
  encodeGetLeadResponse,
  encodeGetLeadStatsResponse,
  encodeImportLeadRowsResponse,
  encodeLeadImportResponse,
  encodeListActivitiesResponse,
  encodeListLeadImportsResponse,
  encodeListLeadSourcesResponse,
  encodeListLeadsResponse,
  encodeListPipelinesResponse,
  encodeListRepsResponse,
  encodeListSalesTasksResponse,
  encodeListStageDealsResponse,
  encodeLogActivityResponse,
  type Lead,
  type Pipeline,
} from '@voxire/voxie/crm/sales/v1/types';
import type { Page, Request } from './fixtures';
import { expect, test } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

/**
 * End-to-end: the CRM sales flow against a stubbed API. Payloads are encoded
 * with the generated contract types, so the wire format matches the server.
 */

const API_URL = 'http://localhost:9999';

const LEADS: Lead[] = [
  {
    id: '11',
    firstName: 'Rana',
    lastName: 'Haddad',
    jobTitle: 'Owner',
    phone: '+961 70 123 456',
    email: 'rana@cedarbakery.test',
    companyName: 'Cedar Bakery',
    source: 'Spring list',
    status: 'attempting',
    assignedTo: '1',
    attemptCount: 2,
    nextFollowUp: '2020-01-01',
    billing: 'one_time',
    createdAt: '1760000000',
  },
  {
    id: '12',
    firstName: 'Karim',
    lastName: 'Saad',
    phone: '+961 3 555 010',
    companyName: 'Saad Motors',
    source: 'Referral',
    status: 'new',
    billing: 'one_time',
    createdAt: '1760000100',
  },
];

const PIPELINE: Pipeline = {
  id: '5',
  name: 'Sales',
  isDefault: true,
  stages: [
    { id: '51', pipelineId: '5', name: 'Qualified', position: 0, probability: 10, kind: 'open' },
    { id: '52', pipelineId: '5', name: 'Proposal', position: 1, probability: 50, kind: 'open' },
    { id: '53', pipelineId: '5', name: 'Won', position: 2, probability: 100, kind: 'won' },
    { id: '54', pipelineId: '5', name: 'Lost', position: 3, probability: 0, kind: 'lost' },
  ],
};

async function stubSales(page: Page): Promise<Request[]> {
  const sent: Request[] = [];
  await page.route(`${API_URL}/crm/**`, (route) => route.fulfill({ json: {} }));
  await stubPortalSession(page, page.context(), {
    isOwner: true,
    appIds: ['crm'],
    permissions: ['crm:customers:view', 'crm:customers:create', 'crm:customers:edit'],
  });
  await page.route(`${API_URL}/hub/subscription/v1/my`, (route) =>
    route.fulfill({ json: { subscription: null } })
  );
  // Matches the path exactly, whatever the query string.
  const json = (path: string, body: unknown) =>
    page.route(
      (url) => url.origin === API_URL && url.pathname === `/crm/sales/v1/${path}`,
      (route) => {
        sent.push(route.request());
        return route.fulfill({ json: body });
      }
    );
  await json('reps', encodeListRepsResponse({ reps: [{ id: '1', name: 'Sam Test' }] }));
  await json(
    'leads/sources',
    encodeListLeadSourcesResponse({ sources: ['Referral', 'Spring list'] })
  );
  await json(
    'leads/stats',
    encodeGetLeadStatsResponse({
      byStatus: [
        { status: 'new', count: 1 },
        { status: 'attempting', count: 1 },
      ],
      followUpsOverdue: 1,
      followUpsToday: 0,
      callsToday: 4,
      connectedToday: 1,
      convertedThisMonth: 0,
    })
  );
  await json('leads', encodeListLeadsResponse({ leads: LEADS, total: 2, nextCursor: '' }));
  await json('leads/11', encodeGetLeadResponse({ lead: LEADS[0] }));
  await json(
    'leads/11/activities',
    encodeListActivitiesResponse({ activities: [], nextCursor: '' })
  );
  await json('tasks', encodeListSalesTasksResponse({ tasks: [], nextCursor: '' }));
  await json('pipelines', encodeListPipelinesResponse({ pipelines: [PIPELINE] }));
  return sent;
}

test.describe('CRM sales', () => {
  test('lists leads with their contact details and groups the sidebar', async ({ page }) => {
    await stubSales(page);
    await page.goto(`${PORTAL_URL}/crm/sales/leads`);

    await expect(page.getByRole('heading', { name: 'Leads', level: 1 })).toBeVisible();
    const row = page.getByRole('row', { name: /Cedar Bakery/ });
    await expect(row).toContainText('Rana Haddad · Owner');
    await expect(row.getByRole('link', { name: '+961 70 123 456' })).toHaveAttribute(
      'href',
      'tel:+96170123456'
    );
    await expect(row).toContainText('Attempting');
    await expect(row).toContainText('Overdue');
    await expect(page.getByRole('tab', { name: /Working/ })).toContainText('2');
    await expect(page.getByText('Showing 2 of 2')).toBeVisible();

    const nav = page.locator('#portal-app-navigation');
    await expect(nav.getByText('Sales', { exact: true })).toBeVisible();
    await expect(nav.getByRole('link', { name: 'Leads' })).toHaveAttribute('aria-current', 'page');
    await page.screenshot({ path: 'test-results/crm-sales-leads.png', fullPage: true });
  });

  test('loads the next page of leads when scrolled to the bottom', async ({ page }) => {
    await stubSales(page);
    const lead = (n: number): Lead => ({
      id: String(100 + n),
      firstName: 'Lead',
      lastName: String(n),
      companyName: `Company ${n}`,
      status: 'new',
      billing: 'one_time',
      createdAt: '1760000000',
    });
    const pages: string[] = [];
    await page.route(
      (url) => url.origin === API_URL && url.pathname === '/crm/sales/v1/leads',
      (route) => {
        const cursor = new URL(route.request().url()).searchParams.get('cursor') ?? '';
        pages.push(cursor);
        const from = cursor === 'p2' ? 60 : 0;
        return route.fulfill({
          json: encodeListLeadsResponse({
            leads: Array.from({ length: 60 }, (_, i) => lead(from + i)),
            total: 120,
            nextCursor: cursor === 'p2' ? '' : 'p2',
          }),
        });
      }
    );
    await page.goto(`${PORTAL_URL}/crm/sales/leads`);

    await expect(page.getByText('Showing 60 of 120')).toBeVisible();
    // The footer sits well below the fold, so nothing more loads until scrolling.
    await page.waitForTimeout(500);
    expect(pages).toEqual(['']);
    await page.getByRole('row', { name: /Company 59/ }).scrollIntoViewIfNeeded();
    await expect(page.getByRole('row', { name: /Company 119/ })).toBeAttached();
    await expect(page.getByText('Showing 120 of 120')).toBeVisible();
    expect(pages).toEqual(['', 'p2']);
    await expect(page.getByRole('button', { name: 'Load more' })).toHaveCount(0);
  });

  test('logs a call with its outcome and a follow-up', async ({ page }) => {
    await stubSales(page);
    let logged: Record<string, unknown> | undefined;
    await page.route(`${API_URL}/crm/sales/v1/activities`, async (route) => {
      logged = route.request().postDataJSON() as Record<string, unknown>;
      await route.fulfill({
        json: encodeLogActivityResponse({
          activity: {
            id: '90',
            leadId: '11',
            type: 'call',
            outcome: 'no_answer',
            occurredAt: '1760000200',
          },
        }),
      });
    });
    await page.goto(`${PORTAL_URL}/crm/sales/leads/11`);

    await expect(page.getByRole('heading', { name: /Cedar Bakery/, level: 1 })).toBeVisible();
    await expect(page.getByRole('link', { name: 'rana@cedarbakery.test' })).toHaveAttribute(
      'href',
      'mailto:rana@cedarbakery.test'
    );
    const panel = page.getByRole('region', { name: 'Log activity' });
    const logCall = panel.getByRole('button', { name: 'Log call' });
    await expect(logCall).toBeDisabled();
    await expect(logCall).toHaveAccessibleDescription('Pick how the call went to log it.');
    await panel.getByRole('button', { name: 'No answer' }).click();
    await expect(logCall).toBeEnabled();
    await expect(panel.getByText('Pick how the call went to log it.')).toHaveCount(0);
    await panel.getByRole('combobox').selectOption('tomorrow');
    await panel.getByRole('button', { name: 'Log call' }).click();

    await expect.poll(() => logged).toBeTruthy();
    expect(logged).toMatchObject({ lead_id: '11', type: 'call', outcome: 'no_answer' });
    expect((logged?.follow_up as Record<string, unknown>)?.due_date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    await page.screenshot({ path: 'test-results/crm-sales-lead.png', fullPage: true });
  });

  test('imports a CSV in chunks with the guessed column mapping', async ({ page }) => {
    await stubSales(page);
    await page.route(`${API_URL}/crm/sales/v1/lead-imports`, (route) =>
      route.request().method() === 'POST'
        ? route.fulfill({
            json: encodeLeadImportResponse({
              import: { id: '77', fileName: 'list.csv', totalRows: 2, status: 'running' },
            }),
          })
        : route.fulfill({ json: encodeListLeadImportsResponse({ imports: [] }) })
    );
    let rows: Record<string, unknown> | undefined;
    await page.route(`${API_URL}/crm/sales/v1/lead-imports/77/rows`, async (route) => {
      rows = route.request().postDataJSON() as Record<string, unknown>;
      await route.fulfill({
        json: encodeImportLeadRowsResponse({
          results: [
            { row: 2, outcome: 'created', leadId: '200' },
            { row: 3, outcome: 'skipped', leadId: '11' },
          ],
        }),
      });
    });
    let finished = false;
    await page.route(`${API_URL}/crm/sales/v1/lead-imports/77`, (route) => {
      finished = true;
      return route.fulfill({
        json: encodeLeadImportResponse({
          import: { id: '77', status: 'completed', created: 1, skipped: 1 },
        }),
      });
    });

    await page.goto(`${PORTAL_URL}/crm/sales/leads/import`);
    await page.getByLabel('Choose a CSV file').setInputFiles({
      name: 'list.csv',
      mimeType: 'text/csv',
      buffer: Buffer.from(
        'Prénom;Nom de famille;Société;Téléphone\nNour;Khalil;Olive Co;+961 1 222 333\nRana;Haddad;Cedar Bakery;+961 70 123 456\n'
      ),
    });

    await expect(page.getByLabel('First name')).toHaveValue('0');
    await expect(page.getByLabel('Company', { exact: true })).toHaveValue('2');
    await page.getByRole('button', { name: 'Import 2 rows' }).click();

    await expect(page.getByText('Import finished')).toBeVisible();
    expect(rows).toMatchObject({
      source: 'list',
      duplicates: 'skip',
      rows: [
        {
          row: 2,
          lead: {
            first_name: 'Nour',
            last_name: 'Khalil',
            company_name: 'Olive Co',
            phone: '+961 1 222 333',
          },
        },
        { row: 3, lead: { company_name: 'Cedar Bakery' } },
      ],
    });
    expect(finished).toBe(true);
    await expect(page.getByText('Duplicates skipped')).toBeVisible();
  });

  test('shows the pipeline as a board with stage totals', async ({ page }) => {
    await stubSales(page);
    await page.route(`${API_URL}/crm/sales/v1/deals/board**`, (route) =>
      route.fulfill({
        json: encodeGetDealBoardResponse({
          pipeline: PIPELINE,
          columns: PIPELINE.stages?.map((stage, index) => ({
            stage,
            deals: index === 1 ? 1 : 0,
            amount: index === 1 ? '6000.00' : '0.00',
            weightedAmount: index === 1 ? '3000.00' : '0.00',
          })),
        }),
      })
    );
    await page.route(`${API_URL}/crm/sales/v1/deals/stats**`, (route) =>
      route.fulfill({
        json: encodeGetDealStatsResponse({
          openDeals: 1,
          openAmount: '6000.00',
          weightedAmount: '3000.00',
          wonThisMonth: 0,
          wonAmountThisMonth: '0.00',
          winRate: 50,
          averageDaysToWin: 12,
          averageWonAmount: '1200.00',
        }),
      })
    );
    await page.route(`${API_URL}/crm/sales/v1/stages/*/deals**`, (route) =>
      route.fulfill({
        json: encodeListStageDealsResponse({
          deals: route.request().url().includes('/stages/52/')
            ? [
                {
                  id: '300',
                  pipelineId: '5',
                  stageId: '52',
                  title: 'Cedar Bakery',
                  contactName: 'Rana Haddad',
                  companyName: 'Cedar Bakery',
                  value: '500.00',
                  billing: 'monthly',
                  amount: '6000.00',
                  status: 'open',
                  assignedTo: '1',
                  createdAt: '1760000000',
                },
              ]
            : [],
          nextCursor: '',
        }),
      })
    );

    await page.goto(`${PORTAL_URL}/crm/sales`);
    await expect(page.getByRole('heading', { name: 'Deals', level: 1 })).toBeVisible();
    const proposal = page.getByRole('region', { name: 'Proposal' });
    await expect(proposal).toContainText('$6,000.00');
    await expect(proposal.getByRole('button', { name: /Cedar Bakery/ })).toContainText('/ mo');
    await expect(page.getByText('50%', { exact: true })).toBeVisible();
    await page.screenshot({ path: 'test-results/crm-sales-board.png', fullPage: true });
  });
});
