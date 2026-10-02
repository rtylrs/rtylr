import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const timeV2 = vi.hoisted(() => ({
  listTeamWorkDays: vi.fn(),
  createWorkDayLine: vi.fn(),
}));
vi.mock('../../shared/lib/clients', () => ({ timeV2 }));

import { fetchTeamWorkDays, useCreateWorkDayLine, useTeamWorkDays } from './useWorkDay';

beforeEach(() => vi.clearAllMocks());

describe('team time queries', () => {
  it('loads every page and propagates a later page failure', async () => {
    timeV2.listTeamWorkDays
      .mockResolvedValueOnce({ workDays: [{ id: '1' }], pagination: { hasNext: true } })
      .mockResolvedValueOnce({ workDays: [{ id: '2' }], pagination: { hasNext: false } });
    expect((await fetchTeamWorkDays('2026-10-01', '2026-10-31')).workDays).toEqual([
      { id: '1' },
      { id: '2' },
    ]);
    expect(timeV2.listTeamWorkDays.mock.calls[1][0].page).toBe(2);
    timeV2.listTeamWorkDays
      .mockResolvedValueOnce({ workDays: [{ id: '1' }], pagination: { hasNext: true } })
      .mockRejectedValueOnce(new Error('Report page failed'));
    await expect(fetchTeamWorkDays('2026-10-01', '2026-10-31')).rejects.toThrow(
      'Report page failed'
    );
  });

  it('refreshes an active team report when a person logs hours', async () => {
    let loggedMinutes = 0;
    timeV2.listTeamWorkDays.mockImplementation(async () => ({
      workDays: [{ id: 'day-1', userId: 'user-1', loggedMinutes }],
    }));
    timeV2.createWorkDayLine.mockImplementation(async () => {
      loggedMinutes = 60;
      return { line: { id: 'line-1' } };
    });
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result, unmount } = renderHook(
      () => ({
        report: useTeamWorkDays('2026-10-02', '2026-10-02'),
        create: useCreateWorkDayLine(),
      }),
      { wrapper }
    );
    await waitFor(() => expect(result.current.report.data?.workDays?.[0].loggedMinutes).toBe(0));
    await act(async () => {
      await result.current.create.mutateAsync({
        date: '2026-10-02',
        workDayId: 'day-1',
        clientKey: 'attempt-1',
        durationMinutes: 60,
        description: 'Verification',
        projectId: 'project-1',
        isBillable: false,
      });
    });
    await waitFor(() => expect(result.current.report.data?.workDays?.[0].loggedMinutes).toBe(60));
    unmount();
    client.clear();
  });
});
