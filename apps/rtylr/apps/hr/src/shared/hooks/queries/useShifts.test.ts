import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockHrTransport = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}));
const mockToast = vi.hoisted(() => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
}));
const mockStaffClient = vi.hoisted(() => ({
  listStaff: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockStaffClient,
  unwrap: vi.fn((val) => val),
}));
vi.mock('../../lib/clients', () => ({ hrTransport: mockHrTransport }));
vi.mock('sonner', () => ({ toast: mockToast }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useCreateShiftAssignment, useGetRoster } from './useShifts';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const assign = async () => {
  const { result } = renderHook(() => useCreateShiftAssignment(), { wrapper: createWrapper() });
  await act(async () => {
    await result.current.mutateAsync({ staffId: 4, shiftId: 9, date: '2026-08-20' });
  });
};

describe('useCreateShiftAssignment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // CreateShiftAssignmentResponse.warnings is advisory: the assignment is still
  // created. Dropping the field would leave the roster silently holding a shift
  // that clashes with the assignee's stated availability.
  it('surfaces each availability warning returned with a created assignment', async () => {
    mockHrTransport.post.mockResolvedValue({
      assignment: { id: '1' },
      warnings: [
        'Jane Doe is unavailable 09:00-12:00 on 2026-08-20',
        'Jane Doe is already assigned to another shift on 2026-08-20',
      ],
    });

    await assign();

    expect(mockToast.success).toHaveBeenCalledWith('shifts.assignmentCreated');
    expect(mockToast.warning).toHaveBeenCalledTimes(2);
    expect(mockToast.warning).toHaveBeenCalledWith(
      'Jane Doe is unavailable 09:00-12:00 on 2026-08-20',
      expect.objectContaining({ duration: 10_000 })
    );
  });

  it('stays quiet when the assignment raised no warnings', async () => {
    mockHrTransport.post.mockResolvedValue({ assignment: { id: '1' }, warnings: [] });

    await assign();

    expect(mockToast.success).toHaveBeenCalledWith('shifts.assignmentCreated');
    expect(mockToast.warning).not.toHaveBeenCalled();
  });

  // A backend that predates the warnings field omits it entirely, which must not
  // throw on the way to the success toast.
  it('tolerates a response with no warnings field', async () => {
    mockHrTransport.post.mockResolvedValue({ assignment: { id: '1' } });

    await assign();

    expect(mockToast.success).toHaveBeenCalledWith('shifts.assignmentCreated');
    expect(mockToast.warning).not.toHaveBeenCalled();
  });
});

describe('useGetRoster', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('adapts flat roster assignments into the employee rows the builder renders', async () => {
    mockHrTransport.get.mockImplementation((path: string) => {
      if (path === '/hr/shifts/roster') {
        return Promise.resolve({
          assignments: [
            {
              id: '12',
              employeeId: '4',
              shiftId: '9',
              date: '2026-08-20',
            },
          ],
          conflicts: [{ employeeId: '4', date: '2026-08-20' }],
        });
      }
      if (path === '/hr/shifts') {
        return Promise.resolve({
          shifts: [
            {
              id: '9',
              name: 'Morning',
              startTime: '09:00',
              endTime: '17:00',
              color: '#10b981',
            },
          ],
        });
      }
      throw new Error(`Unexpected request: ${path}`);
    });
    mockStaffClient.listStaff.mockResolvedValue({
      staff: [{ id: '4', firstName: 'Jane', lastName: 'Doe' }],
    });

    const { result } = renderHook(() => useGetRoster({ weekStart: '2026-08-17' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        staffId: 4,
        staffName: 'Jane Doe',
        assignments: [
          {
            id: 12,
            shiftId: 9,
            shiftName: 'Morning',
            date: '2026-08-20',
            startTime: '09:00',
            endTime: '17:00',
            color: '#10b981',
            conflict: true,
          },
        ],
      },
    ]);
  });
});
