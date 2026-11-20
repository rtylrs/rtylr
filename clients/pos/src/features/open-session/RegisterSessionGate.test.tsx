import React from 'react';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const api = vi.hoisted(() => ({ listSessions: vi.fn(), openSession: vi.fn() }));
const route = vi.hoisted(() => ({ pathname: '/' }));

vi.mock('@voxire/api', () => ({ createOneKitClient: () => api }));
vi.mock('react-router-dom', () => ({ useLocation: () => ({ pathname: route.pathname }) }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import RegisterSessionGate from './RegisterSessionGate';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import { useRegisterSessionStore } from '../../shared/stores/registerSessionStore';
import { useTerminalContextStore } from '../../shared/stores/terminalContextStore';
import { useUserStore } from '../../shared/stores/userStore';

const renderGate = () =>
  render(
    <RegisterSessionGate>
      <div>POS screen</div>
    </RegisterSessionGate>
  );

beforeEach(() => {
  api.listSessions.mockReset();
  api.openSession.mockReset();
  route.pathname = '/';
  useRegisterSessionStore.setState({ status: 'unknown' });
  useUserStore.setState({ userId: '7' });
  useTerminalContextStore.setState({ locationId: 3, terminalId: 1 });
});

afterEach(cleanup);

describe('RegisterSessionGate', () => {
  it('shows the POS when the staff member already has an open session', async () => {
    api.listSessions.mockResolvedValue({
      sessions: [{ id: '42', staffId: '7', status: 'open' }],
    });

    renderGate();

    expect(await screen.findByText('POS screen')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Open register' })).toBeNull();
  });

  it('asks for an opening float, opens the session and then shows the POS', async () => {
    api.listSessions.mockResolvedValue({
      sessions: [{ id: '99', staffId: '8', status: 'open' }],
    });
    api.openSession.mockResolvedValue({ sessionId: '43' });

    renderGate();

    const submit = await screen.findByRole('button', { name: 'Open register' });
    expect(screen.queryByText('POS screen')).toBeNull();

    fireEvent.change(screen.getByLabelText('Opening float'), { target: { value: '75' } });
    fireEvent.click(submit);

    expect(await screen.findByText('POS screen')).toBeTruthy();
    expect(api.openSession).toHaveBeenCalledWith({
      locationId: '3',
      staffId: '7',
      openingFloat: 75,
    });
    expect(useCashDrawerStore.getState().openingFloat).toBe(75);
  });

  it('keeps the cashier on the open step with a message when opening fails', async () => {
    api.listSessions.mockResolvedValue({ sessions: [] });
    api.openSession.mockRejectedValue(new TypeError('Failed to fetch'));

    renderGate();

    fireEvent.click(await screen.findByRole('button', { name: 'Open register' }));

    expect((await screen.findByRole('alert')).textContent).toBe(
      'Could not open the register. Check your connection and try again.'
    );
    expect(screen.queryByText('POS screen')).toBeNull();
  });

  it('does not block the POS when the session list cannot be reached', async () => {
    api.listSessions.mockRejectedValue(new TypeError('Failed to fetch'));

    renderGate();

    expect(await screen.findByText('POS screen')).toBeTruthy();
  });

  it('leaves the day summary readable after the session is closed', async () => {
    route.pathname = '/day-summary';
    api.listSessions.mockResolvedValue({ sessions: [] });

    renderGate();

    expect(await screen.findByText('POS screen')).toBeTruthy();
    await waitFor(() => expect(useRegisterSessionStore.getState().status).toBe('none'));
  });
});
