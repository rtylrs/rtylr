import path from 'node:path';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

const api = vi.hoisted(() => ({
  getBuddySettings: vi.fn(),
  putBuddySettings: vi.fn(),
  listBuddyRoutines: vi.fn(),
  setBuddyRoutineEnabled: vi.fn(),
  deleteBuddyRoutine: vi.fn(),
}));
vi.mock('@voxire/api', () => ({
  createOneKitClient: () => api,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));
import BuddyRoutines from './Buddy';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));
beforeEach(() => {
  vi.resetAllMocks();
  api.listBuddyRoutines.mockResolvedValue({
    routines: [
      {
        id: 4,
        name: 'Gym check-in',
        instructions: 'Ask if I went to the gym',
        everyMinutes: 1440,
        enabled: true,
        nextRunAt: '2026-10-31T17:30:00Z',
      },
    ],
  });
  api.getBuddySettings.mockResolvedValue({
    settings: { enabled: false, instructions: '', intervalMinutes: 60, dailyRunCap: 24 },
  });
  api.putBuddySettings.mockImplementation(async (input) => ({ settings: input }));
  api.setBuddyRoutineEnabled.mockResolvedValue({ routine: {} });
  api.deleteBuddyRoutine.mockResolvedValue({});
});
afterEach(cleanup);

const renderRoutines = (isAdmin = false) => {
  const onClose = vi.fn();
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <BuddyRoutines scope="2:1" isAdmin={isAdmin} onClose={onClose} />
    </QueryClientProvider>
  );
  return onClose;
};

describe('Buddy routines', () => {
  it('lists routines and pauses or deletes them', async () => {
    const user = userEvent.setup();
    window.confirm = vi.fn(() => true);
    renderRoutines();
    expect(await screen.findByText('Gym check-in')).toBeInTheDocument();
    expect(screen.getByText(/Every day/)).toBeInTheDocument();
    await user.click(screen.getByRole('checkbox', { name: 'Run Gym check-in' }));
    expect(api.setBuddyRoutineEnabled).toHaveBeenCalledWith({ id: 4, enabled: false });
    await user.click(screen.getByRole('button', { name: 'Delete Gym check-in' }));
    expect(api.deleteBuddyRoutine).toHaveBeenCalledWith({ id: 4 });
  });

  it('shows support alerts only to admins', async () => {
    renderRoutines(false);
    await screen.findByText('Gym check-in');
    expect(screen.queryByText('Support alerts')).not.toBeInTheDocument();
    expect(api.getBuddySettings).not.toHaveBeenCalled();
    cleanup();
    const user = userEvent.setup();
    renderRoutines(true);
    expect(await screen.findByText('Support alerts')).toBeInTheDocument();
    await user.click(screen.getByRole('checkbox', { name: 'Support alerts' }));
    await user.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(api.putBuddySettings).toHaveBeenCalledWith(expect.objectContaining({ enabled: true }))
    );
  });

  it('closes on Escape', async () => {
    const user = userEvent.setup();
    const onClose = renderRoutines();
    await user.keyboard('{Escape}');
    expect(onClose).toHaveBeenCalled();
  });
});
