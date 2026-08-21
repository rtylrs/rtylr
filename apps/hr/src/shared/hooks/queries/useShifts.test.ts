import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
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

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((val) => val) }));
vi.mock('../../lib/clients', () => ({ hrTransport: mockHrTransport }));
vi.mock('sonner', () => ({ toast: mockToast }));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useCreateShiftAssignment } from './useShifts';

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
