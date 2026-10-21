// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getMilestones: vi.fn(),
    createMilestone: vi.fn(),
    updateMilestone: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useMilestones, useCreateMilestone, useUpdateMilestone } from './useMilestones';
import { flowApi } from '../../shared/services/flowApi';
import { toast } from 'sonner';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children),
    client,
  };
};

describe('useMilestones (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns a raw array response unchanged', async () => {
    const data = [{ id: 'm-1', name: 'Launch' }];
    (flowApi.getMilestones as unknown as Mock).mockResolvedValue(data);
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useMilestones({ projectId: 'p-1' }), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(flowApi.getMilestones).toHaveBeenCalledWith({ projectId: 'p-1' });
  });

  it('unwraps the enveloped { milestones: [...] } shape', async () => {
    const milestones = [{ id: 'm-2', name: 'Beta' }];
    (flowApi.getMilestones as unknown as Mock).mockResolvedValue({ milestones });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useMilestones('p-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(milestones);
  });

  it('fails when the envelope has no milestones key', async () => {
    (flowApi.getMilestones as unknown as Mock).mockResolvedValue({ unrelated: true });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useMilestones(), { wrapper });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data).toBeUndefined();
  });
});

describe('useCreateMilestone', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createMilestone with the given data and invalidates milestones on success', async () => {
    (flowApi.createMilestone as unknown as Mock).mockResolvedValue({ id: 'm-new' });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateMilestone(), { wrapper });
    const input = { projectId: 'p-1', name: 'Launch', targetDate: '2026-08-31' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.createMilestone).toHaveBeenCalledWith(input);
    expect(spy).toHaveBeenCalledWith({ queryKey: ['milestones'] });
  });

  it('shows an error toast and does not invalidate when the create call fails', async () => {
    (flowApi.createMilestone as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateMilestone(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ name: 'Launch' }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('flowToasts.milestone.createFailed');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useUpdateMilestone', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateMilestone with id and data and invalidates milestones on success', async () => {
    (flowApi.updateMilestone as unknown as Mock).mockResolvedValue({});
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateMilestone(), { wrapper });
    await act(() => result.current.mutateAsync({ id: 'm-1', data: { name: 'Renamed' } }));
    expect(flowApi.updateMilestone).toHaveBeenCalledWith('m-1', { name: 'Renamed' });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['milestones'] });
  });

  it('shows an error toast and does not invalidate when the update call fails', async () => {
    (flowApi.updateMilestone as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateMilestone(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'm-1', data: { name: 'Renamed' } }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('flowToasts.milestone.updateFailed');
    expect(spy).not.toHaveBeenCalled();
  });
});
