import React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

// The real board, scan and ticket screens against the in-memory pager. Real
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
// Unlocked unless a test needs the ticket's "Notify me" tap on screen.
const alarmState = vi.hoisted(() => ({ ringing: false, unlocked: true }));
vi.mock('../../../../shared/alarm/useAlarm', () => ({
  useAlarm: () => alarmState,
}));
vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

import Try from './Try';
import { TICKET_POLL_INTERVAL_MS } from '../../shared/lib/ticketPoller';
import { installFakePushBrowser } from '../../shared/testing/fakePushBrowser';
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
  alarmState.unlocked = true;
});

afterEach(() => {
  vi.useRealTimers();
});

describe('pager demo', () => {
  it('scans the demo phone, labels it on the board and rings it on ready', async () => {
    render(<Try />);
    const phone = screen.getByRole('region', { name: "Your customer's phone" });
    const counter = screen.getByRole('region', { name: 'Your counter' });

    // The phone has "scanned": a code and nothing to type.
    const heading = await within(phone).findByRole('heading', { name: /^Your code / });
    // The board names the card by the same number the phone shows.
    const code = heading.querySelector('bdi')?.textContent ?? '';
    expect(code).toMatch(/^[1-9]\d{2}$/);
    expect(within(phone).getByText('Show this number to the cashier')).toBeInTheDocument();
    expect(within(phone).queryByRole('textbox')).not.toBeInTheDocument();

    await nextPoll();
    fireEvent.change(
      await within(counter).findByLabelText(`Order number or name for code ${code}`),
      {
        target: { value: '42' },
      }
    );
    fireEvent.click(within(counter).getByRole('button', { name: 'Save' }));
    await nextPoll();

    expect(await within(phone).findByRole('heading', { name: 'Order 42' })).toBeInTheDocument();
    expect(within(phone).queryByText(code)).not.toBeInTheDocument();
    expect(within(phone).getByText("We're preparing your order")).toBeInTheDocument();
    fireEvent.click(await within(counter).findByRole('button', { name: 'Ready: 42' }));
    await nextPoll();

    expect(await within(phone).findByText('Your order is ready!')).toBeInTheDocument();
    expect(syncSpy).toHaveBeenLastCalledWith(
      expect.objectContaining({ ringKey: expect.any(String) })
    );
  });

  it('never asks for notification permission, even on a browser with Web Push', async () => {
    // Demo tickets exist only in this page, so the server has nothing to push.
    // Any call to the real API, the push key included, throws above.
    const pushBrowser = installFakePushBrowser();
    alarmState.unlocked = false;
    try {
      render(<Try />);
      const phone = screen.getByRole('region', { name: "Your customer's phone" });

      expect(await within(phone).findByText('Show this number to the cashier')).toBeInTheDocument();
      fireEvent.click(within(phone).getByRole('button', { name: "Notify me when it's ready" }));
      await nextPoll();

      expect(pushBrowser.requestPermission).not.toHaveBeenCalled();
      expect(pushBrowser.register).not.toHaveBeenCalled();
      expect(pushBrowser.subscribe).not.toHaveBeenCalled();
    } finally {
      pushBrowser.restore();
    }
  });

  it('lets the visitor change the demo venue settings', async () => {
    render(<Try />);
    const counter = screen.getByRole('region', { name: 'Your counter' });
    const mode = await within(counter).findByLabelText('Label orders by');
    expect(mode).toBeEnabled();
    expect(within(counter).queryByText(/Only owners and admins/)).not.toBeInTheDocument();
  });

  it('gives the demo counter no venue QR, no display and no poster link', async () => {
    render(<Try />);
    const counter = screen.getByRole('region', { name: 'Your counter' });
    await within(counter).findByLabelText('Label orders by');
    expect(within(counter).queryByText('Your pager QR')).not.toBeInTheDocument();
    expect(screen.queryByText('Counter display')).not.toBeInTheDocument();
    expect(screen.queryByText('Print the counter poster')).not.toBeInTheDocument();
  });

  it('starts over with a new code for another phone', async () => {
    render(<Try />);
    const phone = screen.getByRole('region', { name: "Your customer's phone" });
    const first = await within(phone).findByRole('heading', { name: /^Your code / });
    const firstCode = first.textContent;

    fireEvent.click(screen.getByRole('button', { name: 'Scan with another phone' }));

    const second = await within(phone).findByRole('heading', { name: /^Your code / });
    expect(second.textContent).not.toBe(firstCode);
    await nextPoll();
    const counter = screen.getByRole('region', { name: 'Your counter' });
    expect(
      await within(counter).findByRole('heading', { name: 'New phones (2)' })
    ).toBeInTheDocument();
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
