import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  listDepartments: vi.fn(),
  listStaff: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (Client: { name?: string }) =>
    Client.name === 'DepartmentServiceClient'
      ? { listDepartments: mocks.listDepartments }
      : { listStaff: mocks.listStaff },
  getOneKitErrorMessage: (_err: unknown, fallback: string) => fallback,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useDepartments } from './useDepartments';

function createWrapper() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: qc }, children);
}

describe('useDepartments', () => {
  beforeEach(() => {
    mocks.listDepartments.mockReset();
    mocks.listStaff.mockReset();
  });

  it('still returns departments when staff directory access is 403', async () => {
    mocks.listDepartments.mockResolvedValue({
      departments: [
        { id: '12', name: 'Roastery', description: 'Coffee roasting', managerStaffId: '0' },
      ],
    });
    mocks.listStaff.mockRejectedValue(
      Object.assign(new Error('The company directory is turned off for this business'), {
        status: 403,
      })
    );

    const { result } = renderHook(() => useDepartments(false), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 12,
        name: 'Roastery',
        description: 'Coffee roasting',
        staffCount: 0,
        head: '',
        avgSalary: 0,
      },
    ]);
  });
});
