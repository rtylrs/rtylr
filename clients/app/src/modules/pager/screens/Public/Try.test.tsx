import React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

// The real board, join and ticket screens against the in-memory pager. Real
// translations, so the English below also proves the keys resolve.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
  FullScreenLoader: () => <p>Loading</p>,
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({}),
  useNavigate: () => vi.fn(),
}));

// The demo must never reach the real API.
const realApi = vi.hoisted(
  () =>
    new Proxy(
      {},
      {
        get: (_target, method) => () => {
          throw new Error(`The demo called the pager API: ${String(method)}`);
        },
      }
    )
);
vi.mock('../../shared/lib/clients', () => ({
  publicPagerClient: realApi,
  staffPagerClient: realApi,
}));
vi.mock('../../../../shared/lib/clients', () => ({ configClient: realApi }));

const syncSpy = vi.hoisted(() => vi.fn());
vi.mock('../../../../shared/alarm/readyAlarm', () => ({
  readyAlarm: { unlock: vi.fn(), sync: syncSpy, release: vi.fn(), silence: vi.fn() },
}));
vi.mock('../../../../shared/alarm/useAlarm', () => ({
  useAlarm: () => ({ ringing: false, unlocked: true }),
}));
vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

import Try from './Try';
import { TICKET_POLL_INTERVAL_MS } from '../../shared/lib/ticketPoller';
import { initTestI18n } from '../../shared/testing/i18n';

// Both screens poll, as they do against the API.
const nextPoll = (): Promise<void> =>
  act(async () => {
    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS);
  });

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.useFakeTimers({ shouldAdvanceTime: true });
  syncSpy.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('pager demo', () => {
  it('rings the demo phone when the board marks its number ready', async () => {
    render(<Try />);
    const phone = screen.getByRole('region', { name: "Your customer's phone" });
    const counter = screen.getByRole('region', { name: 'Your counter' });

    fireEvent.change(await within(phone).findByLabelText('Order number'), {
      target: { value: '42' },
    });
    fireEvent.click(within(phone).getByRole('button', { name: "Ring me when it's ready" }));
    expect(await within(phone).findByText("We're preparing your order")).toBeInTheDocument();

    await nextPoll();
    fireEvent.click(await within(counter).findByRole('button', { name: 'Ready: order 42' }));
    await nextPoll();

    expect(await within(phone).findByText('Your order is ready!')).toBeInTheDocument();
    expect(syncSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ ringKey: expect.any(String) })
    );
  });

  it('lets the visitor change the demo venue settings', async () => {
    render(<Try />);
    const counter = screen.getByRole('region', { name: 'Your counter' });
    const strict = await within(counter).findByRole('switch', {
      name: 'Only numbers entered here',
    });
    expect(strict).toBeEnabled();
    expect(within(counter).queryByText(/Only owners and admins/)).not.toBeInTheDocument();
  });

  it('gives the demo counter no customer link and no poster link', async () => {
    render(<Try />);
    const counter = screen.getByRole('region', { name: 'Your counter' });
    await within(counter).findByRole('switch', { name: 'Only numbers entered here' });
    expect(within(counter).queryByText('Customer link')).not.toBeInTheDocument();
    expect(screen.queryByText('Print the counter poster')).not.toBeInTheDocument();
  });

  it('sends the visitor to the ordinary sign-up, returning to the pager setup', () => {
    render(<Try />);
    const claim = new URL(
      screen.getByRole('link', { name: 'Get my free pager' }).getAttribute('href') ?? ''
    );
    expect(claim.pathname).toBe('/register');
    expect(claim.searchParams.get('redirect')).toBe('dash');
    expect(claim.searchParams.get('returnTo')).toBe('/pager/setup');

    const signIn = new URL(
      screen.getByRole('link', { name: 'Sign in to switch it on' }).getAttribute('href') ?? ''
    );
    expect(signIn.pathname).toBe('/');
    expect(signIn.searchParams.get('returnTo')).toBe('/pager/setup');
  });
});
