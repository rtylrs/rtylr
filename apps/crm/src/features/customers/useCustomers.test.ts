import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useCustomers,
  useCustomer,
  useCustomerTimeline,
  useCreateCustomer,
  useUpdateCustomer,
  useDeleteCustomer,
  useRecordVisit,
  useImportCustomers,
  useMergeCustomers,
  useExportCustomerData,
} from './useCustomers';

const mockCrmApi = {
  getCustomers: vi.fn(),
  getCustomerDetail: vi.fn(),
  getCustomerTimeline: vi.fn(),
  createCustomer: vi.fn(),
  updateCustomer: vi.fn(),
  deleteCustomer: vi.fn(),
  recordVisit: vi.fn(),
  importCustomers: vi.fn(),
  mergeCustomers: vi.fn(),
  exportCustomerData: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/crmApi', () => ({
  crmApi: {
    getCustomers: (...args: any[]) => mockCrmApi.getCustomers(...args),
    getCustomerDetail: (...args: any[]) => mockCrmApi.getCustomerDetail(...args),
    getCustomerTimeline: (...args: any[]) => mockCrmApi.getCustomerTimeline(...args),
    createCustomer: (...args: any[]) => mockCrmApi.createCustomer(...args),
    updateCustomer: (...args: any[]) => mockCrmApi.updateCustomer(...args),
    deleteCustomer: (...args: any[]) => mockCrmApi.deleteCustomer(...args),
    recordVisit: (...args: any[]) => mockCrmApi.recordVisit(...args),
    importCustomers: (...args: any[]) => mockCrmApi.importCustomers(...args),
    mergeCustomers: (...args: any[]) => mockCrmApi.mergeCustomers(...args),
    exportCustomerData: (...args: any[]) => mockCrmApi.exportCustomerData(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleCustomers = [
  {
    id: 1,
    name: 'Alice',
    email: 'alice@test.com',
    phone: '555-0001',
    totalSpent: 500,
    visits: 10,
    loyaltyPoints: 100,
    lastVisit: '2026-04-01',
    initials: 'AL',
  },
  {
    id: 2,
    name: 'Bob',
    email: 'bob@test.com',
    phone: '555-0002',
    totalSpent: 200,
    visits: 3,
    loyaltyPoints: 50,
    lastVisit: '2026-03-15',
    initials: 'BO',
  },
];

describe('useCustomers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches customers without params', async () => {
    mockCrmApi.getCustomers.mockResolvedValue(sampleCustomers);

    const { result } = renderHook(() => useCustomers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
  });

  it('passes params to getCustomers', async () => {
    mockCrmApi.getCustomers.mockResolvedValue([]);

    const params = { search: 'alice', tier: 'gold' };
    renderHook(() => useCustomers(params), { wrapper: createWrapper() });

    await waitFor(() => expect(mockCrmApi.getCustomers).toHaveBeenCalledWith(params));
  });

  it('handles API error', async () => {
    mockCrmApi.getCustomers.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useCustomers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useCustomer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single customer', async () => {
    mockCrmApi.getCustomerDetail.mockResolvedValue(sampleCustomers[0]);

    const { result } = renderHook(() => useCustomer(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.name).toBe('Alice');
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useCustomer(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCustomerTimeline', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches timeline for a customer', async () => {
    const timeline = [{ id: 1, type: 'visit', title: 'Store visit', at: '2026-01-01' }];
    mockCrmApi.getCustomerTimeline.mockResolvedValue(timeline);

    const { result } = renderHook(() => useCustomerTimeline(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useCustomerTimeline(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreateCustomer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createCustomer with data', async () => {
    mockCrmApi.createCustomer.mockResolvedValue({ id: 3, name: 'Charlie' });

    const { result } = renderHook(() => useCreateCustomer(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ name: 'Charlie', email: 'c@test.com' });
    });

    expect(mockCrmApi.createCustomer).toHaveBeenCalledWith({
      name: 'Charlie',
      email: 'c@test.com',
    });
  });
});

describe('useUpdateCustomer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateCustomer with id and data', async () => {
    mockCrmApi.updateCustomer.mockResolvedValue({ id: 1, name: 'Alice Updated' });

    const { result } = renderHook(() => useUpdateCustomer(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, data: { name: 'Alice Updated' } });
    });

    expect(mockCrmApi.updateCustomer).toHaveBeenCalledWith(1, { name: 'Alice Updated' });
  });
});

describe('useDeleteCustomer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls deleteCustomer with id', async () => {
    mockCrmApi.deleteCustomer.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useDeleteCustomer(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync(1);
    });

    expect(mockCrmApi.deleteCustomer).toHaveBeenCalledWith(1);
  });
});

describe('useRecordVisit', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls recordVisit with customerId', async () => {
    mockCrmApi.recordVisit.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useRecordVisit(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync(1);
    });

    expect(mockCrmApi.recordVisit).toHaveBeenCalledWith(1);
  });
});

describe('useMergeCustomers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls mergeCustomers with primaryId and secondaryId', async () => {
    mockCrmApi.mergeCustomers.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useMergeCustomers(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ primaryId: 1, secondaryId: 2 });
    });

    expect(mockCrmApi.mergeCustomers).toHaveBeenCalledWith(1, 2);
  });
});
