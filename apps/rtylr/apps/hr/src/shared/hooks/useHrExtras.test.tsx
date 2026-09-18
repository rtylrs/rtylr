import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  getSettings: vi.fn(),
  getMyEmployee: vi.fn(),
  listBenefits: vi.fn(),
}));

vi.mock('../services/hrExtras', () => ({
  hrExtras: {
    getSettings: mocks.getSettings,
    getMyEmployee: mocks.getMyEmployee,
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  createOneKitClient: () => ({
    listBenefits: mocks.listBenefits,
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { useHrSettings, useMyEmployeeId, useBenefits } from './useHrExtras';

const settings = {
  workWeekStart: 'Monday' as const,
  overtimeAfterHours: 40,
  ptoPolicyDays: 20,
  currency: 'USD' as const,
  weeklyHours: 40,
};

function createWrapper(): React.FC<{ children: React.ReactNode }> {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe('useHrSettings', () => {
  beforeEach(() => {
    mocks.getSettings.mockReset();
  });

  it('unwraps the /hr/config response envelope before parsing settings', async () => {
    mocks.getSettings.mockResolvedValue({ config: settings });

    const { result } = renderHook(() => useHrSettings(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(settings);
  });
});

describe('useMyEmployeeId', () => {
  beforeEach(() => {
    mocks.getMyEmployee.mockReset();
  });

  it("resolves the caller's own hr_employees.id from /hr/employees/me, not the hub auth user id", async () => {
    mocks.getMyEmployee.mockResolvedValue({ employee: { id: '4201' } });

    const { result } = renderHook(() => useMyEmployeeId(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toBe(4201);
  });
});

describe('useBenefits', () => {
  beforeEach(() => {
    mocks.listBenefits.mockReset();
  });

  it('maps hr.benefit.v1 BenefitPlan fields into the UI shape instead of hitting the dead /hr/benefits REST route', async () => {
    mocks.listBenefits.mockResolvedValue({
      plans: [
        {
          id: '9001',
          ownerId: '42',
          name: 'Premium Health',
          type: 'health',
          provider: 'Acme Insurance',
          costEmployer: 150,
          costEmployee: 25,
          isActive: true,
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
    });

    const { result } = renderHook(() => useBenefits(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: '9001',
        name: 'Premium Health',
        category: 'Health',
        carrier: 'Acme Insurance',
        monthlyCost: 25,
      },
    ]);
    expect(mocks.listBenefits).toHaveBeenCalledWith({});
  });
});
