import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val: unknown) => val),
  createOneKitClient: () => ({ listPayroll: vi.fn(), listStaff: vi.fn() }),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useListPayrollRuns, useTriggerPayrollRun } from './usePayroll';
import { safe } from '../../shared/lib/api';

const createWrapper = (): React.FC<{ children: React.ReactNode }> => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }): React.JSX.Element =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useListPayrollRuns', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches payroll runs on mount', async () => {
    const data = { runs: [], total: 0, limit: 50, offset: 0 };
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(data);
    const { result } = renderHook(() => useListPayrollRuns(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useListPayrollRuns(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useTriggerPayrollRun', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls post with period data', async () => {
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue({});
    const { result } = renderHook(() => useTriggerPayrollRun(), {
      wrapper: createWrapper(),
    });
    await act(() =>
      result.current.mutateAsync({ periodStart: '2026-04-01', periodEnd: '2026-04-30' })
    );
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/runs/trigger', {
      periodStart: '2026-04-01',
      periodEnd: '2026-04-30',
    });
  });
});
