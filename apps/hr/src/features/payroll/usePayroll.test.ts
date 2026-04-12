import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/hrApi', () => ({
  hrApi: {
    listPayroll: vi.fn(),
    runPayroll: vi.fn(),
    markPayrollPaid: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { usePayroll, useRunPayroll, useMarkPayrollPaid } from './usePayroll';
import { hrApi } from '../../shared/services/hrApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('usePayroll', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches payroll list on mount', async () => {
    const data = [
      {
        id: 1,
        name: 'Jane Doe',
        role: 'Manager',
        hoursWorked: 160,
        basePay: 4000,
        overtime: 500,
        deductions: 500,
      },
    ];
    (hrApi.listPayroll as any).mockResolvedValue(data);
    const { result } = renderHook(() => usePayroll(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes params to the service', async () => {
    (hrApi.listPayroll as any).mockResolvedValue([]);
    const params = { period: '2026-04' };
    const { result } = renderHook(() => usePayroll(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(hrApi.listPayroll).toHaveBeenCalledWith(params);
  });

  it('starts in loading state', () => {
    (hrApi.listPayroll as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => usePayroll(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useRunPayroll', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls runPayroll with period', async () => {
    (hrApi.runPayroll as any).mockResolvedValue({ id: 5 });
    const { result } = renderHook(() => useRunPayroll(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('2026-04'));
    expect(hrApi.runPayroll).toHaveBeenCalledWith('2026-04');
  });
});

describe('useMarkPayrollPaid', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls markPayrollPaid with id', async () => {
    (hrApi.markPayrollPaid as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useMarkPayrollPaid(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(7));
    expect(hrApi.markPayrollPaid).toHaveBeenCalledWith(7);
  });
});
