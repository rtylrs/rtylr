import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode }) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

const api = vi.hoisted(() => ({
  listTickets: vi.fn(),
  createTicket: vi.fn(),
  markReady: vi.fn(),
  remindTicket: vi.fn(),
  undoReady: vi.fn(),
  collectTicket: vi.fn(),
  getSettings: vi.fn(),
  updateSettings: vi.fn(),
}));
vi.mock('../../shared/lib/clients', () => ({ staffPagerClient: api }));

const getConfig = vi.hoisted(() => vi.fn());
vi.mock('../../../../shared/lib/clients', () => ({ configClient: { getConfig } }));

import Board from './Board';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useUserStore } from '../../../../shared/stores/userStore';
import { initTestI18n } from '../../shared/testing/i18n';

const tickets = [
  { id: '1', number: '41', status: 'waiting', phoneLinked: false },
  { id: '2', number: '42', status: 'waiting', phoneLinked: true },
  {
    id: '3',
    number: '40',
    status: 'ready',
    phoneLinked: true,
    acknowledgedAt: '2026-11-19T10:05:20.000Z',
  },
];

function renderBoard() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <Board />
    </QueryClientProvider>
  );
}

const section = async (name: RegExp): Promise<HTMLElement> => {
  const heading = await screen.findByRole('heading', { name });
  return heading.closest('section') as HTMLElement;
};

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  Object.values(api).forEach((mock) => mock.mockReset().mockResolvedValue({}));
  api.listTickets.mockResolvedValue({ tickets, serverTime: '2026-11-19T10:06:00.000Z' });
  api.getSettings.mockResolvedValue({ enabled: true, strict: false });
  getConfig.mockReset().mockResolvedValue({ business: { slug: 'acme' } });
  useUserStore.setState({
    user: {
      businessSlug: 'acme',
      effectiveAccess: createEffectiveAccess(undefined, { isOwner: true }),
    },
  });
});

describe('pager staff board', () => {
  it('sorts open tickets into waiting and ready, with who has a phone linked', async () => {
    renderBoard();

    const waiting = await section(/^Waiting \(2\)$/);
    const ready = await section(/^Ready \(1\)$/);

    expect(within(waiting).getByText('41')).toBeInTheDocument();
    expect(within(waiting).getByText('42')).toBeInTheDocument();
    expect(
      within(waiting).getAllByRole('img', { name: 'A phone is following this order' })
    ).toHaveLength(1);
    expect(within(ready).getByText('40')).toBeInTheDocument();
    expect(within(ready).getByText('Customer is on the way')).toBeInTheDocument();
  });

  it('enters a number from the keypad', async () => {
    renderBoard();
    const keypad = await screen.findByRole('group', { name: 'Number keypad' });

    fireEvent.click(within(keypad).getByRole('button', { name: '4' }));
    fireEvent.click(within(keypad).getByRole('button', { name: '3' }));
    fireEvent.click(within(keypad).getByRole('button', { name: '9' }));
    fireEvent.click(within(keypad).getByRole('button', { name: 'Delete the last character' }));
    expect(screen.getByLabelText('Order number')).toHaveValue('43');

    fireEvent.click(screen.getByRole('button', { name: 'Add order' }));

    await waitFor(() => expect(api.createTicket).toHaveBeenCalledWith({ number: '43' }));
    await waitFor(() => expect(screen.getByLabelText('Order number')).toHaveValue(''));
  });

  it.each([
    ['Ready: order 41', 'markReady', '1'],
    ['Remove: order 41', 'collectTicket', '1'],
    ['Call again: order 40', 'remindTicket', '3'],
    ['Undo: order 40', 'undoReady', '3'],
    ['Collected: order 40', 'collectTicket', '3'],
  ] as const)('"%s" calls %s and reloads the board', async (name, method, id) => {
    renderBoard();

    fireEvent.click(await screen.findByRole('button', { name }));

    await waitFor(() => expect(api[method]).toHaveBeenCalledWith({ id }));
    await waitFor(() => expect(api.listTickets).toHaveBeenCalledTimes(2));
  });

  it('switches the venue settings one at a time', async () => {
    api.updateSettings.mockResolvedValue({ enabled: true, strict: true });
    renderBoard();

    const strict = await screen.findByRole('switch', { name: 'Only numbers entered here' });
    await waitFor(() => expect(strict).toBeEnabled());
    expect(screen.getByRole('switch', { name: 'Pager on' })).toBeChecked();
    expect(strict).not.toBeChecked();

    fireEvent.click(strict);

    await waitFor(() => expect(api.updateSettings).toHaveBeenCalledWith({ strict: true }));
    await waitFor(() => expect(strict).toBeChecked());
  });

  it('shows the settings to a member who is not an admin, without letting them change any', async () => {
    useUserStore.setState({
      user: { businessSlug: 'acme', effectiveAccess: createEffectiveAccess('insights') },
    });
    renderBoard();

    const enabled = await screen.findByRole('switch', { name: 'Pager on' });
    await waitFor(() => expect(enabled).toBeChecked());
    expect(enabled).toBeDisabled();
    expect(screen.getByRole('switch', { name: 'Only numbers entered here' })).toBeDisabled();
    expect(
      screen.getByText('Only owners and admins can change these settings.')
    ).toBeInTheDocument();
    // Running the board is still theirs to do.
    expect(await screen.findByRole('button', { name: 'Ready: order 41' })).toBeEnabled();
  });

  it('warns while the pager is switched off', async () => {
    // A false flag is absent from the decoded answer, not false.
    api.getSettings.mockResolvedValue({});
    renderBoard();

    expect(
      await screen.findByText("The pager is off. Customers can't join until you switch it on.")
    ).toBeInTheDocument();
  });

  it('shows the link customers open', async () => {
    renderBoard();

    const link = await screen.findByRole('link', { name: /\/acme\/pager$/ });
    expect(link).toHaveAttribute('href', `${window.location.origin}/acme/pager`);
  });

  it('links to the counter poster', async () => {
    renderBoard();

    expect(await screen.findByRole('link', { name: 'Print the counter poster' })).toHaveAttribute(
      'href',
      '/pager/poster'
    );
  });

  it.each([
    ['too_many_open_tickets', 'Too many open orders. Mark some as collected first.'],
    ['invalid_transition', 'This order has already moved on.'],
    [undefined, "That didn't work. Try again."],
  ])('explains a refused change (%s)', async (code, message) => {
    api.markReady.mockRejectedValue({ statusCode: 409, code });
    renderBoard();

    fireEvent.click(await screen.findByRole('button', { name: 'Ready: order 41' }));

    expect(await screen.findByText(message)).toBeInTheDocument();
  });

  it('tells a member without permission instead of showing an empty board', async () => {
    api.listTickets.mockRejectedValue({ statusCode: 403 });
    renderBoard();

    expect(await screen.findByText("You don't have access to the pager.")).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Add order' })).not.toBeInTheDocument();
  });
});
