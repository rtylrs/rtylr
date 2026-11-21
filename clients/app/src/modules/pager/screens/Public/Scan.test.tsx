import React, { StrictMode } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

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
const scanned = { publicToken: 'tok-1', code: 'A7', color: 2, status: 'waiting' };

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  window.sessionStorage.clear();
  navigateSpy.mockReset();
  api.scanTicket.mockReset().mockResolvedValue(scanned);
  api.getPublicTicket.mockReset();
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
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(1);
    expect(api.scanTicket).toHaveBeenCalledWith({ slug: 'acme' });
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(window.sessionStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('keeps the same ticket and code when the page is reloaded', async () => {
    window.sessionStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue({ status: 'waiting', code: 'K3' });
    render(<Scan />);

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-old', replace: true })
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
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true })
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
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-old', replace: true })
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
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true })
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
    [{ statusCode: 500 }, "We couldn't reach the pager. Check your connection and try again."],
  ])('explains a failed scan and retries on a tap: %j', async (failure, message) => {
    api.scanTicket.mockRejectedValueOnce(failure);
    render(<Scan />);

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(navigateSpy).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(2);
  });
});
