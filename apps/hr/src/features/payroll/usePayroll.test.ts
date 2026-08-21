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

vi.mock('../../shared/lib/clients', () => ({
  hrTransport: {
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
  useApprovePayrollRun,
} from './usePayroll';
import { useListMyPayrollRuns, useGetMyPayrollRun } from '../../shared/hooks/queries/usePayroll';
import { hrTransport } from '../../shared/lib/clients';

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
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue(data);
    const { result } = renderHook(() => useListPayrollRuns(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
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
    (hrTransport.post as ReturnType<typeof vi.fn>).mockResolvedValue({});
    const { result } = renderHook(() => useTriggerPayrollRun(), {
      wrapper: createWrapper(),
    });
    await act(() =>
      result.current.mutateAsync({ periodStart: '2026-04-01', periodEnd: '2026-04-30' })
    );
    expect(hrTransport.post).toHaveBeenCalledWith('/hr/payroll/runs/trigger', {
      periodStart: '2026-04-01',
      periodEnd: '2026-04-30',
    });
  });
});

describe('payroll configuration query and action gates', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('does not fetch any payroll run when configuration is unavailable', () => {
    const { result: adminList } = renderHook(
      () => useListPayrollRuns(undefined, { enabled: false }),
      { wrapper: createWrapper() }
    );
    const { result: adminDetail } = renderHook(() => useGetPayrollRun(7, { enabled: false }), {
      wrapper: createWrapper(),
    });
    const { result: selfList } = renderHook(
      () => useListMyPayrollRuns(undefined, { enabled: false }),
      { wrapper: createWrapper() }
    );
    const { result: selfDetail } = renderHook(() => useGetMyPayrollRun(7, { enabled: false }), {
      wrapper: createWrapper(),
    });

    expect(adminList.current.fetchStatus).toBe('idle');
    expect(adminDetail.current.fetchStatus).toBe('idle');
    expect(selfList.current.fetchStatus).toBe('idle');
    expect(selfDetail.current.fetchStatus).toBe('idle');
    expect(hrTransport.get).not.toHaveBeenCalled();
  });

  it('does not post payroll actions when configuration is unavailable', async () => {
    const { result: trigger } = renderHook(() => useTriggerPayrollRun(false), {
      wrapper: createWrapper(),
    });
    const { result: approve } = renderHook(() => useApprovePayrollRun(false), {
      wrapper: createWrapper(),
    });

    await act(async () => {
      await trigger.current.mutateAsync({ periodStart: '2026-04-01', periodEnd: '2026-04-30' });
      await approve.current.mutateAsync({ runId: 7 });
    });

    expect(hrTransport.post).not.toHaveBeenCalled();
  });
});

describe('useListMyPayrollRuns', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('reads the self-service endpoint, not the admin one', async () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      runs: [],
      total: 0,
      limit: 50,
      offset: 0,
    });
    const { result } = renderHook(() => useListMyPayrollRuns({ limit: 50, offset: 0 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(hrTransport.get).toHaveBeenCalledWith('/hr/payroll/my-runs', {
      params: { limit: 50, offset: 0 },
    });
  });

  it('sends no status param since the server already restricts to approved/exported', async () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      runs: [],
      total: 0,
      limit: 50,
      offset: 0,
    });
    const { result } = renderHook(() => useListMyPayrollRuns({ limit: 50, offset: 0 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const [, config] = (hrTransport.get as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(config.params).not.toHaveProperty('status');
  });

  it('does not share a cache entry with the admin list hook', async () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({
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

    const urls = (hrTransport.get as ReturnType<typeof vi.fn>).mock.calls.map(([url]) => url);
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
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run });
    const { result } = renderHook(() => useGetMyPayrollRun(7), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(hrTransport.get).toHaveBeenCalledWith('/hr/payroll/my-runs/7');
    expect(result.current.data).toEqual(run);
  });

  it('does not fetch when no run id is given', () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run: {} });
    renderHook(() => useGetMyPayrollRun(undefined), { wrapper: createWrapper() });
    expect(hrTransport.get).not.toHaveBeenCalled();
  });

  it('does not share a cache entry with the admin detail hook', async () => {
    (hrTransport.get as ReturnType<typeof vi.fn>).mockResolvedValue({ run: { id: 7 } });
    const client = createClient();
    const wrapper = wrapperFor(client);
    const admin = renderHook(() => useGetPayrollRun(7), { wrapper });
    const self = renderHook(() => useGetMyPayrollRun(7), { wrapper });

    await waitFor(() => expect(admin.result.current.isSuccess).toBe(true));
    await waitFor(() => expect(self.result.current.isSuccess).toBe(true));

    const urls = (hrTransport.get as ReturnType<typeof vi.fn>).mock.calls.map(([url]) => url);
    expect(urls).toContain('/hr/payroll/runs/7');
    expect(urls).toContain('/hr/payroll/my-runs/7');
    expect(client.getQueryCache().getAll()).toHaveLength(2);
  });
});
