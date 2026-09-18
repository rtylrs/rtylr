// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getCycles: vi.fn(),
    createCycle: vi.fn(),
    updateCycle: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: vi.fn(() => ({ activeWorkspaceId: 'ws-store' })),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useCycles, useCreateCycle, useUpdateCycle } from './useCycles';
import { flowApi } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';
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

describe('useCycles (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (useFlowStore as unknown as Mock).mockReturnValue({ activeWorkspaceId: 'ws-store' });
  });

  // Cycles are project-scoped on v2, so the id is the project's and there is no
  // store fallback. These two used to cover a workspace fallback and an
  // explicit override of it.
  it("asks for the given project's cycles", async () => {
    (flowApi.getCycles as unknown as Mock).mockResolvedValue([{ id: 'c-1' }]);
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useCycles('p-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(flowApi.getCycles).toHaveBeenCalledWith('p-1');
  });

  // v2 declares project_id @required @gt(0), so firing without one would 400.
  it('does not fetch until a project is known', () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useCycles(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getCycles).not.toHaveBeenCalled();
  });

  it('unwraps the enveloped { cycles: [...] } shape', async () => {
    const cycles = [{ id: 'c-2', name: 'Sprint 1' }];
    (flowApi.getCycles as unknown as Mock).mockResolvedValue({ cycles });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useCycles('p-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(cycles);
  });
});

describe('useCreateCycle', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (useFlowStore as unknown as Mock).mockReturnValue({ activeWorkspaceId: 'ws-store' });
  });

  // Creating a cycle carries the project through untouched; nothing is stamped
  // on from a store any more.
  it('sends the payload through untouched and invalidates cycles', async () => {
    (flowApi.createCycle as unknown as Mock).mockResolvedValue({ id: 'c-new' });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateCycle(), { wrapper });
    await act(() => result.current.mutateAsync({ projectId: 'p-1', name: 'Sprint 1' }));
    expect(flowApi.createCycle).toHaveBeenCalledWith({ projectId: 'p-1', name: 'Sprint 1' });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['cycles'] });
  });

  it('shows an error toast and does not invalidate when the create call fails', async () => {
    (flowApi.createCycle as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateCycle(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ name: 'Sprint 1' }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('toasts.cycle.createFailed');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useUpdateCycle', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateCycle with id and data and invalidates cycles on success', async () => {
    (flowApi.updateCycle as unknown as Mock).mockResolvedValue({});
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateCycle(), { wrapper });
    await act(() => result.current.mutateAsync({ id: 'c-1', data: { name: 'Renamed' } }));
    expect(flowApi.updateCycle).toHaveBeenCalledWith('c-1', { name: 'Renamed' });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['cycles'] });
  });

  it('shows an error toast and does not invalidate when the update call fails', async () => {
    (flowApi.updateCycle as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateCycle(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'c-1', data: { name: 'Renamed' } }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('toasts.cycle.updateFailed');
    expect(spy).not.toHaveBeenCalled();
  });
});
