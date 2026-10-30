import path from 'node:path';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

const api = vi.hoisted(() => ({
  getBuddySettings: vi.fn(),
  putBuddySettings: vi.fn(),
  listBuddyRuns: vi.fn(),
  listBuddyActions: vi.fn(),
  decideBuddyAction: vi.fn(),
}));
vi.mock('@voxire/api', () => ({
  createOneKitClient: () => api,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));
import Buddy from './Buddy';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));
beforeEach(() => {
  vi.resetAllMocks();
  api.getBuddySettings.mockResolvedValue({
    settings: {
      enabled: false,
      instructions: 'Check tickets',
      intervalMinutes: 60,
      dailyRunCap: 24,
    },
  });
  api.putBuddySettings.mockImplementation(async (input) => ({ settings: input }));
  api.listBuddyRuns.mockResolvedValue({
    runs: [{ id: 1, status: 'done', report: 'Everything is quiet.' }],
  });
  api.listBuddyActions.mockResolvedValue({
    actions: [{ id: 8, toolName: 'update_ticket', input: { id: 12 } }],
  });
  api.decideBuddyAction.mockResolvedValue({ action: { status: 'executed' } });
});
afterEach(cleanup);
function showBuddy() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={queryClient}>
      <Buddy scope="1:1" onBack={() => {}} />
    </QueryClientProvider>
  );
}
describe('Caret Buddy', () => {
  it('saves edited settings and approves a proposal using the generated contract', async () => {
    const user = userEvent.setup();
    showBuddy();
    await screen.findByDisplayValue('Check tickets');
    await user.click(screen.getByRole('checkbox', { name: 'Enable Buddy' }));
    await user.clear(screen.getByLabelText('Interval (minutes)'));
    await user.type(screen.getByLabelText('Interval (minutes)'), '30');
    await user.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(api.putBuddySettings).toHaveBeenCalledWith({
        enabled: true,
        instructions: 'Check tickets',
        intervalMinutes: 30,
        dailyRunCap: 24,
      })
    );
    expect(screen.getByText('Everything is quiet.')).toBeInTheDocument();
    expect(screen.getByText(/"id": 12/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Approve' }));
    await waitFor(() =>
      expect(api.decideBuddyAction).toHaveBeenCalledWith({ actionId: 8, decision: 'approve' })
    );
    expect(await screen.findByText('Action decision saved.')).toBeInTheDocument();
  });
  it('shows a failed execution and surfaces list failures rather than empty states', async () => {
    api.listBuddyRuns.mockRejectedValue(new Error('offline'));
    api.decideBuddyAction.mockResolvedValue({
      action: { status: 'failed', result: { error: 'unavailable' } },
    });
    const user = userEvent.setup();
    showBuddy();
    expect(await screen.findByText('Unable to load recent runs.')).toBeInTheDocument();
    await user.click(await screen.findByRole('button', { name: 'Approve' }));
    expect(
      await screen.findByText(/The approved action could not be completed:/)
    ).toHaveTextContent('unavailable');
  });
});
