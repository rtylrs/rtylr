import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockPayrollClient = vi.hoisted(() => ({
  listPayroll: vi.fn(),
}));
const mockStaffClient = vi.hoisted(() => ({
  listStaff: vi.fn(),
}));

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val: unknown) => val),
  createOneKitClient: (Client: { name: string }) =>
    Client.name === 'PayrollServiceClient' ? mockPayrollClient : mockStaffClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import {
  usePayroll,
  useListPayrollRuns,
  useGetPayrollRun,
  useTriggerPayrollRun,
} from './usePayroll';
import { useListMyPayrollRuns, useGetMyPayrollRun } from '../../shared/hooks/queries/usePayroll';
import { safe } from '../../shared/lib/api';

const createClient = (): QueryClient =>
  new QueryClient({ defaultOptions: { queries: { retry: false } } });

const wrapperFor =
  (client: QueryClient): React.FC<{ children: React.ReactNode }> =>
  ({ children }: { children: React.ReactNode }): React.JSX.Element =>
    React.createElement(QueryClientProvider, { client }, children);

const createWrapper = (): React.FC<{ children: React.ReactNode }> => wrapperFor(createClient());

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

describe('usePayroll compensation gate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not request payroll or staff compensation data when disabled', () => {
    const { result } = renderHook(() => usePayroll({ enabled: false }), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockPayrollClient.listPayroll).not.toHaveBeenCalled();
    expect(mockStaffClient.listStaff).not.toHaveBeenCalled();
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

describe('useListMyPayrollRuns', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('reads the self-service endpoint, not the admin one', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      runs: [],
      total: 0,
      limit: 50,
      offset: 0,
    });
    const { result } = renderHook(() => useListMyPayrollRuns({ limit: 50, offset: 0 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/hr/payroll/my-runs', {
      params: { limit: 50, offset: 0 },
    });
  });

  it('sends no status param since the server already restricts to approved/exported', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      runs: [],
      total: 0,
      limit: 50,
      offset: 0,
    });
    const { result } = renderHook(() => useListMyPayrollRuns({ limit: 50, offset: 0 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const [, config] = (safe.get as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(config.params).not.toHaveProperty('status');
  });

  it('does not share a cache entry with the admin list hook', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      runs: [],
      total: 0,
      limit: 50,
      offset: 0,
    });
    const client = createClient();
    const wrapper = wrapperFor(client);
    const admin = renderHook(() => useListPayrollRuns({ limit: 50, offset: 0 }), { wrapper });
    const self = renderHook(() => useListMyPayrollRuns({ limit: 50, offset: 0 }), { wrapper });

    await waitFor(() => expect(admin.result.current.isSuccess).toBe(true));
    await waitFor(() => expect(self.result.current.isSuccess).toBe(true));

    const urls = (safe.get as ReturnType<typeof vi.fn>).mock.calls.map(([url]) => url);
    expect(urls).toContain('/hr/payroll/runs');
    expect(urls).toContain('/hr/payroll/my-runs');
    expect(client.getQueryCache().getAll()).toHaveLength(2);
  });
});

describe('useGetMyPayrollRun', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('reads the self-service run endpoint and unwraps run', async () => {
    const run = { id: 7, status: 'exported', entries: [{ employeeId: 42, netPay: 1000 }] };
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run });
    const { result } = renderHook(() => useGetMyPayrollRun(7), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/hr/payroll/my-runs/7');
    expect(result.current.data).toEqual(run);
  });

  it('does not fetch when no run id is given', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run: {} });
    renderHook(() => useGetMyPayrollRun(undefined), { wrapper: createWrapper() });
    expect(safe.get).not.toHaveBeenCalled();
  });

  it('does not share a cache entry with the admin detail hook', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run: { id: 7 } });
    const client = createClient();
    const wrapper = wrapperFor(client);
    const admin = renderHook(() => useGetPayrollRun(7), { wrapper });
    const self = renderHook(() => useGetMyPayrollRun(7), { wrapper });

    await waitFor(() => expect(admin.result.current.isSuccess).toBe(true));
    await waitFor(() => expect(self.result.current.isSuccess).toBe(true));

    const urls = (safe.get as ReturnType<typeof vi.fn>).mock.calls.map(([url]) => url);
    expect(urls).toContain('/hr/payroll/runs/7');
    expect(urls).toContain('/hr/payroll/my-runs/7');
    expect(client.getQueryCache().getAll()).toHaveLength(2);
  });
});
