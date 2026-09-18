import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockStaffClient = vi.hoisted(() => ({
  listStaff: vi.fn(),
}));
const mockDepartmentClient = vi.hoisted(() => ({
  listDepartments: vi.fn(),
}));

vi.mock('../../shared/services/hrApi', () => ({
  hrApi: {
    createStaff: vi.fn(),
    updateStaff: vi.fn(),
    deleteStaff: vi.fn(),
    clockIn: vi.fn(),
    clockOut: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
  createOneKitClient: (Client: { name: string }) => {
    switch (Client.name) {
      case 'StaffServiceClient':
        return mockStaffClient;
      case 'DepartmentServiceClient':
        return mockDepartmentClient;
      default:
        throw new Error(`unexpected client ${Client.name}`);
    }
  },
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

import {
  useStaff,
  useCreateStaff,
  useUpdateStaff,
  useDeleteStaff,
  useClockIn,
  useClockOut,
} from './useStaff';
import { hrApi } from '../../shared/services/hrApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useStaff', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches staff list and resolves department names', async () => {
    mockStaffClient.listStaff.mockResolvedValue({
      staff: [
        {
          id: '1',
          firstName: 'Jane',
          lastName: 'Doe',
          email: 'jane@example.com',
          role: 'Manager',
          departmentId: 'd1',
          employmentType: 'Full-time',
          status: 'Active',
        },
      ],
    });
    mockDepartmentClient.listDepartments.mockResolvedValue({
      departments: [{ id: 'd1', name: 'Sales' }],
    });

    const { result } = renderHook(() => useStaff(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([
      {
        id: 1,
        name: 'Jane Doe',
        role: 'Manager',
        department: 'Sales',
        email: 'jane@example.com',
        employmentType: 'Full-time',
        status: 'Active',
      },
    ]);
  });

  it('starts in loading state', () => {
    mockStaffClient.listStaff.mockReturnValue(new Promise(() => {}));
    mockDepartmentClient.listDepartments.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useStaff(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useCreateStaff', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createStaff service method', async () => {
    (hrApi.createStaff as unknown as Mock).mockResolvedValue({ id: 2 });
    const { result } = renderHook(() => useCreateStaff(), {
      wrapper: createWrapper(),
    });
    const input: {
      name: string;
      email: string;
      role: string;
      department: string;
      employmentType: 'Full-time' | 'Part-time' | 'Contractor';
    } = {
      name: 'John',
      email: 'john@example.com',
      role: 'Staff',
      department: 'Ops',
      employmentType: 'Full-time',
    };
    await act(() => result.current.mutateAsync(input));
    expect(hrApi.createStaff).toHaveBeenCalledWith(input);
  });
});

describe('useUpdateStaff', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateStaff with data', async () => {
    (hrApi.updateStaff as unknown as Mock).mockResolvedValue({ id: 1 });
    const { result } = renderHook(() => useUpdateStaff(), {
      wrapper: createWrapper(),
    });
    const input = { id: 1, name: 'Jane Updated' };
    await act(() => result.current.mutateAsync(input));
    expect(hrApi.updateStaff).toHaveBeenCalledWith(input);
  });
});

describe('useDeleteStaff', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteStaff with id', async () => {
    (hrApi.deleteStaff as unknown as Mock).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteStaff(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(3));
    expect(hrApi.deleteStaff).toHaveBeenCalledWith(3);
  });
});

describe('useClockIn', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls clockIn with staffId', async () => {
    (hrApi.clockIn as unknown as Mock).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useClockIn(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(1));
    expect(hrApi.clockIn).toHaveBeenCalledWith(1);
  });
});

describe('useClockOut', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls clockOut with staffId', async () => {
    (hrApi.clockOut as unknown as Mock).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useClockOut(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync(1));
    expect(hrApi.clockOut).toHaveBeenCalledWith(1);
  });
});
