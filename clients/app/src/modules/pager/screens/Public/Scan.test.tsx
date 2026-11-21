import React, { StrictMode } from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

// Real translations: an assertion written in English also proves the key resolves.

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'Acme' }),
  useNavigate: () => navigateSpy,
}));

const api = vi.hoisted(() => ({ scanTicket: vi.fn(), getPublicTicket: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ publicPagerClient: api }));

import Scan from './Scan';
import { initTestI18n } from '../../shared/testing/i18n';

const STORAGE_KEY = 'pager:scan:acme';
const scanned = { publicToken: 'tok-1', word: 'panda', color: 2, status: 'waiting' };

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  window.sessionStorage.clear();
  navigateSpy.mockReset();
  api.scanTicket.mockReset().mockResolvedValue(scanned);
  api.getPublicTicket.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('pager scan page', () => {
  it('opens a ticket on load, with no tap and nothing to type, then shows it', async () => {
    // StrictMode runs effects twice in development: still one ticket.
    render(
      <StrictMode>
        <Scan />
      </StrictMode>
    );

    expect(screen.getByRole('status')).toHaveTextContent('Getting your code...');
    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(1);
    expect(api.scanTicket).toHaveBeenCalledWith({ slug: 'acme' });
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(window.sessionStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('keeps the same ticket and code when the page is reloaded', async () => {
    window.sessionStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue({ status: 'waiting', word: 'slay' });
    render(<Scan />);

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-old', replace: true })
    );
    expect(api.getPublicTicket).toHaveBeenCalledWith({ publicToken: 'tok-old' });
    expect(api.scanTicket).not.toHaveBeenCalled();
  });

  it.each([
    ['expired', { status: 'expired' }],
    ['collected', { status: 'collected' }],
  ])('scans again when the remembered ticket is %s', async (_, ticket) => {
    window.sessionStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue(ticket);
    render(<Scan />);

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(window.sessionStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('scans again when the remembered ticket is gone', async () => {
    window.sessionStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockRejectedValue({ statusCode: 404, code: 'ticket_not_found' });
    render(<Scan />);

    await waitFor(() => expect(api.scanTicket).toHaveBeenCalledTimes(1));
    expect(window.sessionStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('keeps the remembered ticket when it cannot be checked, instead of a second phone', async () => {
    window.sessionStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockRejectedValue({ statusCode: 503 });
    render(<Scan />);

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-old', replace: true })
    );
    expect(api.scanTicket).not.toHaveBeenCalled();
  });

  it('still works when the browser refuses session storage', async () => {
    const getItem = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    const setItem = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    render(<Scan />);

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    getItem.mockRestore();
    setItem.mockRestore();
  });

  it('hands the ticket to the demo instead of navigating, and remembers nothing', async () => {
    const onScanned = vi.fn();
    render(<Scan slug="demo" onScanned={onScanned} />);

    await waitFor(() => expect(onScanned).toHaveBeenCalledWith('tok-1'));
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it('says so when the venue has no pager', async () => {
    api.scanTicket.mockRejectedValue({ statusCode: 404, code: 'venue_not_found' });
    render(<Scan />);

    expect(await screen.findByText("This pager isn't available")).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it.each([
    [{ statusCode: 429 }, 'Too many attempts. Wait a moment, then try again.'],
    [
      { statusCode: 429, code: 'too_many_open_tickets' },
      'This venue has too many phones waiting right now. Please ask at the counter.',
    ],
  ])('explains a failed scan and retries on a tap: %j', async (failure, message) => {
    api.scanTicket.mockRejectedValueOnce(failure);
    render(<Scan />);

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(navigateSpy).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(2);
  });

  // While a new version deploys, the API answers with nothing, a 5xx or an
  // uncoded 404 for a few seconds.
  it('rides out a deploy calmly, retrying after 1, 2 and 4 seconds', async () => {
    vi.useFakeTimers();
    api.scanTicket
      .mockRejectedValueOnce({ statusCode: 502 })
      .mockRejectedValueOnce({ statusCode: 404 })
      .mockRejectedValueOnce(new TypeError('Failed to fetch'))
      .mockResolvedValue(scanned);
    render(<Scan />);

    await act(() => vi.advanceTimersByTimeAsync(1_000));
    expect(screen.getByRole('status')).toHaveTextContent('Still getting your code, one moment...');
    await act(() => vi.advanceTimersByTimeAsync(2_000));
    expect(navigateSpy).not.toHaveBeenCalled();
    await act(() => vi.advanceTimersByTimeAsync(4_000));

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true });
    expect(api.scanTicket).toHaveBeenCalledTimes(4);
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows the error with a retry once the deploy has taken about 20 seconds', async () => {
    vi.useFakeTimers();
    api.scanTicket.mockRejectedValue({ statusCode: 503 });
    render(<Scan />);

    await act(() => vi.advanceTimersByTimeAsync(14_000));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    await act(() => vi.advanceTimersByTimeAsync(1_000));

    expect(screen.getByRole('alert')).toHaveTextContent(
      "We couldn't reach the pager. Check your connection and try again."
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(5);

    api.scanTicket.mockResolvedValue(scanned);
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    await act(() => vi.advanceTimersByTimeAsync(0));
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true });
  });

  it('does not retry a venue without a pager', async () => {
    vi.useFakeTimers();
    api.scanTicket.mockRejectedValue({ statusCode: 404, code: 'venue_not_found' });
    render(<Scan />);

    await act(() => vi.advanceTimersByTimeAsync(20_000));
    expect(api.scanTicket).toHaveBeenCalledTimes(1);
  });
});
