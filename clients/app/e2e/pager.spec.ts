import { test, expect, type Page } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

/**
 * Pager e2e. The customer pages are public and mounted on the Portal's root
 * route, so they need no auth cookie; the staff board sits inside the
 * signed-in shell.
 *
 * What only a browser can show: the join tap and the later ring happen in the
 * same document, with the bell unlocked by that tap. A full page load between
 * the two would throw the unlocked audio element away.
 */

type TicketState = 'waiting' | 'ready' | 'acknowledged';

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

const bellPlays = (page: Page): Promise<BellPlay[]> => page.evaluate(() => window.__bellPlays);
const audiblePlays = async (page: Page): Promise<number> =>
  (await bellPlays(page)).filter((play) => !play.muted && play.outcome === 'played').length;

async function stubPublicPager(page: Page): Promise<{ ticket: TicketState }> {
  const state: { ticket: TicketState } = { ticket: 'waiting' };

  await page.route('**/pager/v1/public/venues/acme', (route) =>
    route.fulfill({ json: { name: 'Acme Cafe', enabled: true, strict: false } })
  );
  await page.route('**/pager/v1/public/venues/acme/join', (route) =>
    route.fulfill({ json: { public_token: 'tok-e2e', number: '42', status: 'waiting' } })
  );
  await page.route('**/pager/v1/public/tickets/tok-e2e', (route) =>
    route.fulfill({
      json: {
        number: '42',
        status: state.ticket === 'waiting' ? 'waiting' : 'ready',
        venue_name: 'Acme Cafe',
        ahead: state.ticket === 'waiting' ? 2 : 0,
        ...(state.ticket === 'waiting' ? {} : { ready_at: '2026-11-19T10:05:00.000Z' }),
        ...(state.ticket === 'acknowledged' ? { acknowledged_at: '2026-11-19T10:05:20.000Z' } : {}),
        server_time: new Date().toISOString(),
      },
    })
  );
  await page.route('**/pager/v1/public/tickets/tok-e2e/acknowledge', (route) => {
    state.ticket = 'acknowledged';
    return route.fulfill({ json: {} });
  });

  return state;
}

test.describe('Pager (customer)', () => {
  test('the join tap unlocks the bell, which rings in the same document once the order is ready', async ({
    page,
  }) => {
    await recordBellPlays(page);
    const state = await stubPublicPager(page);

    await page.goto(`${PORTAL_URL}/acme/pager`);
    await expect(page.getByText('Acme Cafe')).toBeVisible();
    await page.evaluate(() => {
      window.__pagerDocument = 'join';
    });

    const number = page.getByLabel('Order number');
    await expect(number).toHaveAttribute('inputmode', 'numeric');
    await number.fill('42');
    await page.getByRole('button', { name: "Ring me when it's ready" }).click();

    await expect(page).toHaveURL(`${PORTAL_URL}/pager/t/tok-e2e`);
    await expect(page.getByRole('heading', { name: 'Order 42' })).toBeVisible();
    await expect(page.getByText("We're preparing your order")).toBeVisible();
    await expect(page.getByText('2 orders ahead of you')).toBeVisible();

    // Still the document the tap happened in, with the bell primed silently.
    expect(await page.evaluate(() => window.__pagerDocument)).toBe('join');
    await expect
      .poll(async () => (await bellPlays(page))[0])
      .toEqual({ muted: true, outcome: 'played' });
    expect(await audiblePlays(page)).toBe(0);
    await expect(page.getByRole('button', { name: 'Alert me' })).toHaveCount(0);

    state.ticket = 'ready';

    await expect(page.getByRole('alert')).toContainText('Your order is ready!');
    await expect.poll(() => audiblePlays(page)).toBeGreaterThan(0);
    await expect(page.getByRole('button', { name: 'Stop sound' })).toBeVisible();

    await page.getByRole('button', { name: 'I have it' }).click();

    await expect(page.getByText("Thanks. The counter knows you're on your way.")).toBeVisible();
    expect(state.ticket).toBe('acknowledged');
    await expect(page.getByRole('button', { name: 'Stop sound' })).toHaveCount(0);

    // The bell repeats about every two seconds while ringing; it must not now.
    const rungSoFar = await audiblePlays(page);
    await page.waitForTimeout(3_000);
    expect(await audiblePlays(page)).toBe(rungSoFar);
  });

  test('a ticket opened without the join tap asks for one', async ({ page }) => {
    await recordBellPlays(page);
    await stubPublicPager(page);

    await page.goto(`${PORTAL_URL}/pager/t/tok-e2e`);

    await expect(page.getByRole('heading', { name: 'Order 42' })).toBeVisible();
    const alertMe = page.getByRole('button', { name: 'Alert me' });
    await expect(alertMe).toBeVisible();

    await alertMe.click();

    await expect(alertMe).toHaveCount(0);
    await expect
      .poll(async () => (await bellPlays(page))[0])
      .toEqual({ muted: true, outcome: 'played' });
  });

  test('a venue without a pager shows a dead end, not a form', async ({ page }) => {
    await page.route('**/pager/v1/public/venues/nowhere', (route) =>
      route.fulfill({
        status: 404,
        json: { code: 'venue_not_found', message: 'venue not found' },
      })
    );

    await page.goto(`${PORTAL_URL}/nowhere/pager`);

    await expect(page.getByText("This pager isn't available")).toBeVisible();
    await expect(page.getByLabel('Order number')).toHaveCount(0);
  });
});

test.describe('Pager (staff board)', () => {
  test('a member who is not an admin marks a waiting order ready', async ({ page, context }) => {
    await stubPortalSession(page, context);
    const tickets = [
      { id: '1', number: '41', status: 'waiting', source: 'staff', phone_linked: true },
      { id: '2', number: '42', status: 'waiting', source: 'customer', phone_linked: true },
    ];
    const readyCalls: string[] = [];

    await page.route('**/pager/v1/tickets', (route) =>
      route.fulfill({ json: { tickets, server_time: new Date().toISOString() } })
    );
    await page.route('**/pager/v1/tickets/*/ready', (route) => {
      const id = route.request().url().split('/').at(-2) ?? '';
      readyCalls.push(id);
      const ticket = tickets.find((candidate) => candidate.id === id);
      if (ticket) ticket.status = 'ready';
      return route.fulfill({ json: { ticket } });
    });
    await page.route('**/pager/v1/settings', (route) =>
      route.fulfill({ json: { enabled: true, strict: false } })
    );

    await page.goto(`${PORTAL_URL}/pager`);

    await expect(page.getByRole('heading', { name: 'Waiting (2)' })).toBeVisible();
    await expect(page.getByRole('switch', { name: 'Pager on' })).toBeChecked();
    await expect(page.getByRole('switch', { name: 'Pager on' })).toBeDisabled();

    await page.getByRole('button', { name: 'Ready: order 41' }).click();

    await expect(page.getByRole('heading', { name: 'Ready (1)' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Call again: order 41' })).toBeVisible();
    expect(readyCalls).toEqual(['1']);
  });
});
