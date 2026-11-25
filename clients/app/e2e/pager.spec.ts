import { test, expect, type BrowserContext, type Page } from './fixtures';
import { stubPortalSession } from './portal-session';

const RENNILI_URL = process.env.RENNILI_URL ?? 'http://localhost:5198';

/**
 * Pager e2e, on Rennili's own dev server. The customer pages are public, so
 * they need no auth cookie; the staff board needs the signed-in business, read
 * from the same token cookie the Portal uses.
 *
 * What only a browser can show: the QR lands on one upfront tap, which opens
 * the ticket and unlocks the bell; that tap and the later ring happen in the
 * same document; and a reload or a second scan keeps the same ticket.
 *
 * Public Rennili pages open in Arabic unless a language was chosen on Rennili,
 * so the public specs that read English copy choose English first via the
 * host-only cookie the language switch sets. Staff pages default to English.
 */

type TicketState = 'scanned' | 'waiting' | 'ready' | 'acknowledged';

interface BellPlay {
  muted: boolean;
  outcome: string;
}

declare global {
  interface Window {
    __bellPlays: BellPlay[];
    __pagerDocument?: string;
  }
}

/** Records every `play()` on an audio element, and whether the browser allowed it. */
async function recordBellPlays(page: Page): Promise<void> {
  await page.addInitScript(() => {
    const plays: BellPlay[] = [];
    window.__bellPlays = plays;
    const play = HTMLMediaElement.prototype.play;
    HTMLMediaElement.prototype.play = function recordedPlay(this: HTMLMediaElement) {
      const entry = { muted: this.muted, outcome: 'pending' };
      plays.push(entry);
      const result = play.call(this);
      result.then(
        () => {
          entry.outcome = 'played';
        },
        (error: Error) => {
          entry.outcome = error.name;
        }
      );
      return result;
    };
  });
}

/** The host-only cookie Rennili's switch sets; without it public pages open in Arabic. */
async function chooseEnglish(context: BrowserContext): Promise<void> {
  await context.addCookies([{ name: 'rennili_language', value: 'en', url: RENNILI_URL }]);
}

const bellPlays = (page: Page): Promise<BellPlay[]> => page.evaluate(() => window.__bellPlays);
const audiblePlays = async (page: Page): Promise<number> =>
  (await bellPlays(page)).filter((play) => !play.muted && play.outcome === 'played').length;

async function stubPublicPager(page: Page): Promise<{ ticket: TicketState; scans: number }> {
  const state: { ticket: TicketState; scans: number } = { ticket: 'scanned', scans: 0 };

  // The scan page reads the venue before it offers the tap. No push key: the
  // bell on the page is what these specs follow.
  await page.route('**/pager/v1/public/venues/acme', (route) =>
    route.fulfill({ json: { name: 'Acme Cafe', enabled: true } })
  );
  await page.route('**/pager/v1/public/push-key', (route) =>
    route.fulfill({ json: { public_key: '' } })
  );
  await page.route('**/pager/v1/public/venues/acme/scan', (route) => {
    state.scans += 1;
    return route.fulfill({
      json: { public_token: 'tok-e2e', code: '427', status: 'waiting' },
    });
  });
  await page.route('**/pager/v1/public/tickets/tok-e2e', (route) => {
    const labelled = state.ticket !== 'scanned';
    const waiting = state.ticket === 'scanned' || state.ticket === 'waiting';
    return route.fulfill({
      json: {
        ...(labelled ? { label: '42', label_kind: 'number' } : { code: '427' }),
        status: waiting ? 'waiting' : 'ready',
        venue_name: 'Acme Cafe',
        ...(state.ticket === 'waiting' ? { ahead: 2 } : {}),
        ...(waiting ? {} : { ready_at: '2026-11-19T10:05:00.000Z' }),
        ...(state.ticket === 'acknowledged' ? { acknowledged_at: '2026-11-19T10:05:20.000Z' } : {}),
        server_time: new Date().toISOString(),
      },
    });
  });
  await page.route('**/pager/v1/public/tickets/tok-e2e/acknowledge', (route) => {
    state.ticket = 'acknowledged';
    return route.fulfill({ json: {} });
  });

  return state;
}

test.describe('Pager (landing)', () => {
  test('renders the landing page, toggles FAQ items and reaches /try', async ({
    page,
    context,
  }) => {
    await chooseEnglish(context);
    await page.goto(`${RENNILI_URL}/`);

    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();

    // Open a closed FAQ accordion item by keyboard.
    const faqTrigger = page.locator('#faq button[aria-controls="rl-faq-panel-1"]');
    await expect(faqTrigger).toHaveAttribute('aria-expanded', 'false');
    await faqTrigger.focus();
    await page.keyboard.press('Enter');
    await expect(faqTrigger).toHaveAttribute('aria-expanded', 'true');

    // Dedicated route reaches the RTYLR services upsell page on both desktop and mobile.
    await page.getByRole('link', { name: 'Explore RTYLR services' }).click();
    await expect(page).toHaveURL(`${RENNILI_URL}/rtylr`);
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();

    // Primary CTA reaches the live demo without calling any real API.
    await page.locator('main').getByRole('link', { name: 'Try the live demo' }).first().click();
    await expect(page).toHaveURL(`${RENNILI_URL}/try`);
    await expect(page.getByRole('region', { name: "Your customer's phone" })).toBeVisible();
  });
});

test.describe('Pager (customer)', () => {
  test('the one tap on the scan page opens a code with nothing to type and unlocks the bell, which rings once the order is ready', async ({
    page,
    context,
  }) => {
    await chooseEnglish(context);
    await recordBellPlays(page);
    const state = await stubPublicPager(page);

    await page.goto(`${RENNILI_URL}/acme`);

    // The QR lands on a single tap, with nothing to type and no ticket yet.
    const start = page.getByRole('button', { name: 'Get my code & enable alerts' });
    await expect(start).toBeVisible();
    await expect(page.getByText('Acme Cafe')).toBeVisible();
    await expect(page.getByRole('textbox')).toHaveCount(0);
    expect(state.scans).toBe(0);
    expect(await bellPlays(page)).toEqual([]);

    await page.evaluate(() => {
      window.__pagerDocument = 'scan';
    });
    await start.click();

    // That tap opened the ticket and primed the bell silently.
    await expect(page).toHaveURL(`${RENNILI_URL}/t/tok-e2e`);
    await expect(page.getByRole('heading', { name: 'Your code 427' })).toBeVisible();
    await expect(page.getByText('Show this number to the cashier')).toBeVisible();
    await expect(page.getByRole('textbox')).toHaveCount(0);
    expect(state.scans).toBe(1);
    await expect
      .poll(async () => (await bellPlays(page))[0])
      .toEqual({ muted: true, outcome: 'played' });
    expect(await audiblePlays(page)).toBe(0);
    // No second ask on the ticket: the alerts are already on.
    await expect(page.getByText('Alerts active')).toBeVisible();
    await expect(page.getByRole('button', { name: "Notify me when it's ready" })).toHaveCount(0);

    // The cashier labels the phone.
    state.ticket = 'waiting';
    await expect(page.getByRole('heading', { name: 'Order 42' })).toBeVisible();
    await expect(page.getByText("We're preparing your order")).toBeVisible();
    await expect(page.getByText('2 orders ahead of you')).toBeVisible();

    state.ticket = 'ready';

    await expect(page.getByRole('alert')).toContainText('Your order is ready!');
    await expect.poll(() => audiblePlays(page)).toBeGreaterThan(0);
    await expect(page.getByRole('button', { name: 'Stop sound' })).toBeVisible();
    // Still the document the tap happened in.
    expect(await page.evaluate(() => window.__pagerDocument)).toBe('scan');

    await page.getByRole('button', { name: 'I have it' }).click();

    await expect(page.getByText("Thanks. The counter knows you're on your way.")).toBeVisible();
    expect(state.ticket).toBe('acknowledged');
    await expect(page.getByRole('button', { name: 'Stop sound' })).toHaveCount(0);

    // The bell repeats about every two seconds while ringing; it must not now.
    const rungSoFar = await audiblePlays(page);
    await page.waitForTimeout(3_000);
    expect(await audiblePlays(page)).toBe(rungSoFar);
  });

  test('a reload or a second scan keeps the same ticket, and a reloaded page offers the tap again', async ({
    page,
    context,
  }) => {
    await chooseEnglish(context);
    await recordBellPlays(page);
    const state = await stubPublicPager(page);

    await page.goto(`${RENNILI_URL}/acme`);
    await page.getByRole('button', { name: 'Get my code & enable alerts' }).click();
    await expect(page).toHaveURL(`${RENNILI_URL}/t/tok-e2e`);
    expect(state.scans).toBe(1);

    // A reload keeps the same ticket and code.
    await page.reload();
    await expect(page.getByRole('heading', { name: 'Your code 427' })).toBeVisible();

    // A reload is a new document with the bell locked again, so the page asks
    // for the one tap here; it unlocks the bell without ringing out loud.
    const notify = page.getByRole('button', { name: "Notify me when it's ready" });
    await expect(notify).toBeVisible();
    await notify.click();
    await expect
      .poll(async () => (await bellPlays(page))[0])
      .toEqual({ muted: true, outcome: 'played' });
    expect(await audiblePlays(page)).toBe(0);
    await expect(notify).toHaveCount(0);
    await expect(page.getByText('Alerts active')).toBeVisible();

    // Scanning the QR again resumes the same ticket rather than adding a phone.
    await page.goto(`${RENNILI_URL}/acme`);
    await expect(page.getByRole('button', { name: 'Start a new order instead' })).toBeVisible();
    await page.getByRole('button', { name: 'Get my code & enable alerts' }).click();
    await expect(page).toHaveURL(`${RENNILI_URL}/t/tok-e2e`);
    await expect(page.getByRole('heading', { name: 'Your code 427' })).toBeVisible();
    expect(state.scans).toBe(1);
  });

  test('a venue without a pager shows a dead end', async ({ page, context }) => {
    await chooseEnglish(context);
    let scans = 0;
    await page.route('**/pager/v1/public/venues/nowhere', (route) =>
      route.fulfill({
        status: 404,
        json: { code: 'venue_not_found', message: 'venue not found' },
      })
    );
    await page.route('**/pager/v1/public/venues/nowhere/scan', (route) => {
      scans += 1;
      return route.fulfill({
        status: 404,
        json: { code: 'venue_not_found', message: 'venue not found' },
      });
    });

    await page.goto(`${RENNILI_URL}/nowhere`);

    await expect(page.getByText("This pager isn't available")).toBeVisible();
    await expect(page.getByRole('textbox')).toHaveCount(0);
    // No tap to offer and nothing put on the board.
    await expect(page.getByRole('button', { name: 'Get my code & enable alerts' })).toHaveCount(0);
    expect(scans).toBe(0);
  });
});

/**
 * The count on a queue's heading. A narrow board shows one queue at a time
 * behind tabs, so that queue's tab is picked first, as a cashier would.
 */
async function queueCount(page: Page, name: string) {
  await page.locator('.pager-board-workspace').waitFor();
  const tab = page.getByRole('tab', { name });
  if (await tab.isVisible()) await tab.click();
  return page.getByRole('heading', { name }).locator('.pager-board-count');
}

test.describe('Pager (staff board)', () => {
  test('a member who is not an admin labels a new phone and marks it ready', async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context);
    const tickets: Record<string, unknown>[] = [
      { id: '1', label: '41', label_kind: 'number', status: 'waiting' },
      { id: '2', code: '427', status: 'waiting', created_at: new Date().toISOString() },
    ];
    const labelCalls: unknown[] = [];
    const readyCalls: string[] = [];

    await page.route('**/pager/v1/tickets', (route) =>
      route.fulfill({ json: { tickets, server_time: new Date().toISOString() } })
    );
    await page.route('**/pager/v1/tickets/*/label', (route) => {
      const id = route.request().url().split('/').at(-2) ?? '';
      const body = route.request().postDataJSON() as { label: string };
      labelCalls.push({ id, ...body });
      const ticket = tickets.find((candidate) => candidate.id === id);
      if (ticket)
        Object.assign(ticket, { label: body.label, label_kind: 'number', code: undefined });
      return route.fulfill({ json: { ticket } });
    });
    await page.route('**/pager/v1/tickets/*/ready', (route) => {
      const id = route.request().url().split('/').at(-2) ?? '';
      readyCalls.push(id);
      const ticket = tickets.find((candidate) => candidate.id === id);
      if (ticket) ticket.status = 'ready';
      return route.fulfill({ json: { ticket } });
    });
    await page.route('**/pager/v1/settings', (route) =>
      route.fulfill({ json: { enabled: true, label_mode: 'number' } })
    );

    await page.goto(`${RENNILI_URL}/board`);
    await expect(page.locator('html')).toHaveAttribute('lang', 'en');

    // Rennili's own header, not the Portal shell.
    await expect(page.getByRole('link', { name: 'Rennili' }).first()).toBeVisible();
    await expect(page.getByRole('link', { name: 'Open RTYLR' })).toBeVisible();

    await expect(await queueCount(page, 'Waiting')).toHaveText('1');
    await expect(await queueCount(page, 'New phones')).toHaveText('1');
    await expect(page.getByText('Pager live')).toBeVisible();

    // Settings sit in a drawer; a member sees them but cannot change them.
    await page.getByRole('button', { name: 'Settings', exact: true }).click();
    await expect(page.getByRole('switch', { name: 'Pager on' })).toBeChecked();
    await expect(page.getByRole('switch', { name: 'Pager on' })).toBeDisabled();
    await expect(page.getByText('Only owners and admins can change these settings.')).toBeVisible();

    await expect(page.getByRole('button', { name: 'Select code 427' })).toHaveAttribute(
      'aria-pressed',
      'true'
    );
    const label = page.getByLabel('Order number for code 427');
    await expect(label).toHaveAttribute('inputmode', 'numeric');
    await label.fill('42');
    await page.getByRole('button', { name: 'Save', exact: true }).click();

    await expect(await queueCount(page, 'New phones')).toHaveText('0');
    await expect(page.getByText('No phones waiting.')).toBeVisible();
    await expect(await queueCount(page, 'Waiting')).toHaveText('2');
    expect(labelCalls).toEqual([{ id: '2', label: '42' }]);

    await page.getByRole('button', { name: 'Ready: 42' }).click();

    await expect(await queueCount(page, 'Ready')).toHaveText('1');
    await expect(page.getByRole('button', { name: 'Call again: 42' })).toBeVisible();
    expect(readyCalls).toEqual(['2']);
  });
});
