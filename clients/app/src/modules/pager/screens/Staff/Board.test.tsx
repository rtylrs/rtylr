import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode }) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

const qrValues = vi.hoisted(() => [] as string[]);
vi.mock('react-qrcode-logo', () => ({
  QRCode: ({ value }: { value: string }) => {
    qrValues.push(value);
    return <canvas data-testid="qr" />;
  },
}));

const api = vi.hoisted(() => ({
  listTickets: vi.fn(),
  labelTicket: vi.fn(),
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
import { RENNILI_URL } from '@voxire/config';
import { TICKET_POLL_INTERVAL_MS } from '../../shared/lib/ticketPoller';
import { initTestI18n } from '../../shared/testing/i18n';

const newPhone = {
  id: '5',
  code: '427',
  status: 'waiting',
  createdAt: '2026-11-19T10:03:00.000Z',
};
const tickets = [
  { id: '1', label: '41', labelKind: 'number', status: 'waiting' },
  { id: '2', label: 'Sara', labelKind: 'name', status: 'waiting' },
  {
    id: '3',
    label: '40',
    labelKind: 'number',
    status: 'ready',
    acknowledgedAt: '2026-11-19T10:05:20.000Z',
  },
  newPhone,
];
const serverTime = '2026-11-19T10:06:00.000Z';

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
  qrValues.length = 0;
  Object.values(api).forEach((mock) => mock.mockReset().mockResolvedValue({}));
  api.listTickets.mockResolvedValue({ tickets, serverTime });
  api.getSettings.mockResolvedValue({ enabled: true, labelMode: 'either' });
  getConfig.mockReset().mockResolvedValue({ business: { slug: 'acme' } });
  useUserStore.setState({
    user: {
      businessSlug: 'acme',
      effectiveAccess: createEffectiveAccess(undefined, { isOwner: true }),
    },
  });
});

afterEach(() => {
  vi.useRealTimers();
});

describe('pager staff board', () => {
  it('sorts open tickets into new phones, waiting and ready', async () => {
    renderBoard();

    const fresh = await section(/^New phones \(1\)$/);
    const waiting = await section(/^Waiting \(2\)$/);
    const ready = await section(/^Ready \(1\)$/);

    expect(within(fresh).getByText('427')).toBeInTheDocument();
    expect(within(fresh).getByText('Scanned 3 min ago')).toBeInTheDocument();
    expect(within(waiting).getByText('41')).toBeInTheDocument();
    expect(within(waiting).getByText('Sara')).toBeInTheDocument();
    expect(within(ready).getByText('40')).toBeInTheDocument();
    expect(within(ready).getByText('Customer is on the way')).toBeInTheDocument();
  });

  it('shows a phone that scans within one poll', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    api.listTickets.mockResolvedValueOnce({ tickets: [], serverTime });
    api.listTickets.mockResolvedValue({ tickets: [newPhone], serverTime });
    renderBoard();
    await section(/^New phones \(0\)$/);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS);
    });

    const fresh = await section(/^New phones \(1\)$/);
    expect(within(fresh).getByText('427')).toBeInTheDocument();
  });

  // The cashier types an order number on a card that shows a code number, so
  // the two must never read alike.
  it('captions the code apart from the order number it is labelled with', async () => {
    renderBoard();
    const fresh = await section(/^New phones \(1\)$/);
    const code = within(fresh).getByText('427');
    expect(code.previousElementSibling).toHaveTextContent('Code');
    expect(code).toHaveAttribute('dir', 'ltr');
    // The input is captioned with what to type, and has no example number in it.
    const input = within(fresh).getByLabelText('Order number or name for code 427');
    expect(within(fresh).getByText('Order number or name')).toBeVisible();
    expect(input).not.toHaveAttribute('placeholder');

    // Once labelled, the order number is the card's text, captioned Order.
    const waiting = await section(/^Waiting \(2\)$/);
    expect(within(waiting).getByText('41').previousElementSibling).toHaveTextContent('Order');
    expect(within(waiting).queryByText('427')).not.toBeInTheDocument();
  });

  it('still shows the code of a phone that scanned before the numbers', async () => {
    api.listTickets.mockResolvedValue({
      tickets: [{ ...newPhone, code: 'PANDA2' }],
      serverTime,
    });
    renderBoard();
    const fresh = await section(/^New phones \(1\)$/);
    expect(within(fresh).getByText('PANDA2')).toBeInTheDocument();
    expect(within(fresh).getByLabelText('Order number or name for code PANDA2')).toBeEnabled();
  });

  it('labels a new phone with what the cashier types', async () => {
    renderBoard();
    const input = await screen.findByLabelText('Order number or name for code 427');
    expect(input).toHaveAttribute('inputmode', 'text');

    fireEvent.change(input, { target: { value: ' Sara ' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(api.labelTicket).toHaveBeenCalledWith({ id: '5', label: 'Sara' }));
    await waitFor(() => expect(input).toHaveValue(''));
    await waitFor(() => expect(api.listTickets).toHaveBeenCalledTimes(2));
  });

  it('opens a numeric keyboard and shapes the entry as a receipt number in number mode', async () => {
    api.getSettings.mockResolvedValue({ enabled: true, labelMode: 'number' });
    renderBoard();
    const input = await screen.findByLabelText('Order number for code 427');
    await waitFor(() => expect(input).toHaveAttribute('inputmode', 'numeric'));

    fireEvent.change(input, { target: { value: '٤٢' } });

    expect(input).toHaveValue('42');
  });

  it('says so when another open order has the number', async () => {
    api.labelTicket.mockRejectedValue({ statusCode: 409, code: 'label_taken' });
    renderBoard();
    const input = await screen.findByLabelText('Order number or name for code 427');

    fireEvent.change(input, { target: { value: '41' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(
      await screen.findByText('Another open order already has this number.')
    ).toBeInTheDocument();
    expect(input).toHaveValue('41');
  });

  it.each([
    ['Ready: 41', 'markReady', '1'],
    ['Remove: 41', 'collectTicket', '1'],
    ['Remove phone 427', 'collectTicket', '5'],
    ['Call again: 40', 'remindTicket', '3'],
    ['Undo: 40', 'undoReady', '3'],
    ['Collected: 40', 'collectTicket', '3'],
  ] as const)('"%s" calls %s and reloads the board', async (name, method, id) => {
    renderBoard();

    fireEvent.click(await screen.findByRole('button', { name }));

    await waitFor(() => expect(api[method]).toHaveBeenCalledWith({ id }));
    await waitFor(() => expect(api.listTickets).toHaveBeenCalledTimes(2));
  });

  it('changes what orders are labelled by', async () => {
    api.updateSettings.mockResolvedValue({ enabled: true, labelMode: 'name' });
    renderBoard();

    const mode = await screen.findByRole('combobox', { name: 'Label orders by' });
    await waitFor(() => expect(mode).toBeEnabled());
    expect(mode).toHaveTextContent('Order number or name');
    expect(screen.queryByText('Only numbers entered here')).not.toBeInTheDocument();

    fireEvent.click(mode);
    fireEvent.click(screen.getByRole('option', { name: 'Customer name' }));

    await waitFor(() => expect(api.updateSettings).toHaveBeenCalledWith({ labelMode: 'name' }));
    await waitFor(() => expect(mode).toHaveTextContent('Customer name'));
  });

  it('shows the settings to a member who is not an admin, without letting them change any', async () => {
    useUserStore.setState({
      user: { businessSlug: 'acme', effectiveAccess: createEffectiveAccess('insights') },
    });
    renderBoard();

    const enabled = await screen.findByRole('switch', { name: 'Pager on' });
    await waitFor(() => expect(enabled).toBeChecked());
    expect(enabled).toBeDisabled();
    expect(screen.getByLabelText('Label orders by')).toBeDisabled();
    expect(
      screen.getByText('Only owners and admins can change these settings.')
    ).toBeInTheDocument();
    // Running the board is still theirs to do.
    expect(await screen.findByRole('button', { name: 'Ready: 41' })).toBeEnabled();
    expect(screen.getByLabelText('Order number or name for code 427')).toBeEnabled();
  });

  it('warns while the pager is switched off', async () => {
    // A false flag is absent from the decoded answer, not false.
    api.getSettings.mockResolvedValue({});
    renderBoard();

    expect(
      await screen.findByText("The pager is off. Customers can't scan until you switch it on.")
    ).toBeInTheDocument();
  });

  it('shows the venue QR and the link it opens', async () => {
    renderBoard();

    const link = await screen.findByRole('link', { name: /\/acme$/ });
    expect(link).toHaveAttribute('href', `${RENNILI_URL}/acme`);
    expect(qrValues).toContain(`${RENNILI_URL}/acme`);
  });

  it('links to the counter display and the poster', async () => {
    renderBoard();

    expect(await screen.findByRole('link', { name: 'Counter display' })).toHaveAttribute(
      'href',
      '/display'
    );
    expect(screen.getByRole('link', { name: 'Print the counter poster' })).toHaveAttribute(
      'href',
      '/poster'
    );
  });

  it.each([
    ['invalid_transition', 'This order has already moved on.'],
    [undefined, "That didn't work. Try again."],
  ])('explains a refused change (%s)', async (code, message) => {
    api.markReady.mockRejectedValue({ statusCode: 409, code });
    renderBoard();

    fireEvent.click(await screen.findByRole('button', { name: 'Ready: 41' }));

    expect(await screen.findByText(message)).toBeInTheDocument();
  });

  it('tells a member without permission instead of showing an empty board', async () => {
    api.listTickets.mockRejectedValue({ statusCode: 403 });
    renderBoard();

    expect(await screen.findByText("You don't have access to the pager.")).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: /New phones/ })).not.toBeInTheDocument();
  });

  describe('ready-time estimates', () => {
    it('presets the venue time on a new phone and lets the cashier pick another', async () => {
      api.getSettings.mockResolvedValue({ enabled: true, labelMode: 'either', prepMinutes: 8 });
      renderBoard();

      const newPhones = await section(/^New phones/);
      const times = await within(newPhones).findByRole('group', { name: 'Ready in' });
      expect(within(times).getByRole('button', { name: '8 min' })).toHaveAttribute(
        'aria-pressed',
        'true'
      );
      fireEvent.click(within(times).getByRole('button', { name: '15 min' }));
      fireEvent.change(within(newPhones).getByLabelText(/for code 427/), {
        target: { value: '43' },
      });
      fireEvent.click(within(newPhones).getByRole('button', { name: 'Save' }));

      await waitFor(() =>
        expect(api.labelTicket).toHaveBeenCalledWith({ id: '5', label: '43', prepMinutes: 15 })
      );
    });

    it('follows the venue time when it loads after the phone appears', async () => {
      let loadSettings: (value: object) => void = () => undefined;
      api.getSettings.mockReturnValue(
        new Promise((resolve) => {
          loadSettings = resolve;
        })
      );
      renderBoard();
      const newPhones = await section(/^New phones/);
      await within(newPhones).findByLabelText(/for code 427/);

      loadSettings({ enabled: true, labelMode: 'either', prepMinutes: 10 });
      const times = await within(newPhones).findByRole('group', { name: 'Ready in' });
      expect(within(times).getByRole('button', { name: '10 min' })).toHaveAttribute(
        'aria-pressed',
        'true'
      );
    });

    it('shows no time picker when the venue gives no estimate', async () => {
      renderBoard();
      const newPhones = await section(/^New phones/);
      await within(newPhones).findByLabelText(/for code 427/);
      expect(within(newPhones).queryByRole('group', { name: 'Ready in' })).not.toBeInTheDocument();
    });

    it('shows how long is left on a waiting order', async () => {
      api.listTickets.mockResolvedValue({
        tickets: [
          {
            id: '1',
            label: '41',
            labelKind: 'number',
            status: 'waiting',
            expectedReadyAt: '2026-11-19T10:09:00.000Z',
          },
        ],
        serverTime,
      });
      renderBoard();

      expect(await screen.findByText('Due in 3 min')).toBeInTheDocument();
    });
  });

  it('saves the venue review link and turns on typed receipt numbers', async () => {
    api.updateSettings.mockImplementation(async (patch: object) => ({
      enabled: true,
      labelMode: 'either',
      ...patch,
    }));
    renderBoard();

    const link = await screen.findByLabelText('Google review link');
    await waitFor(() => expect(link).toBeEnabled());
    fireEvent.change(link, { target: { value: 'https://g.page/r/acme/review' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save link' }));
    await waitFor(() =>
      expect(api.updateSettings).toHaveBeenCalledWith({
        reviewUrl: 'https://g.page/r/acme/review',
      })
    );

    const numbers = screen.getByRole('switch', { name: 'Customers type their receipt number' });
    await waitFor(() => expect(numbers).toBeEnabled());
    fireEvent.click(numbers);
    await waitFor(() => expect(api.updateSettings).toHaveBeenCalledWith({ customerNumbers: true }));
  });

  it("saves the venue's ringtone", async () => {
    api.updateSettings.mockImplementation(async (patch: object) => ({
      enabled: true,
      labelMode: 'either',
      ringtone: 'bell',
      ...patch,
    }));
    renderBoard();

    const tone = await screen.findByRole('combobox', { name: 'Ringtone' });
    await waitFor(() => expect(tone).toBeEnabled());
    expect(tone).toHaveTextContent('Bell');
    fireEvent.click(tone);
    fireEvent.click(screen.getByRole('option', { name: 'Chime' }));

    await waitFor(() => expect(api.updateSettings).toHaveBeenCalledWith({ ringtone: 'chime' }));
    await waitFor(() => expect(tone).toHaveTextContent('Chime'));
  });
});
