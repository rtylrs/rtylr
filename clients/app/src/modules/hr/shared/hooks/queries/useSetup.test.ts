import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  workSchedule: { createWorkSchedule: vi.fn() },
  department: { createDepartment: vi.fn() },
}));

vi.mock('../../lib/clients', () => ({
  hrTransport: { get: mocks.get, post: mocks.post, put: mocks.put },
}));
vi.mock('../../stores/hrConfigStore', () => ({ useHRConfigStore: vi.fn() }));
vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    ...mocks.workSchedule,
    ...mocks.department,
  }),
  unwrap: (value: unknown) => value,
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { useGetSetupStatus, useCompleteSetupStep } from './useSetup';

const createWrapper = (): React.FC<{ children: React.ReactNode }> => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }) => React.createElement(QueryClientProvider, { client }, children);
};

describe('useCompleteSetupStep', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.post.mockResolvedValue({});
    mocks.put.mockResolvedValue({ config: {} });
    mocks.workSchedule.createWorkSchedule.mockResolvedValue({ schedule: { id: '9' } });
    mocks.department.createDepartment.mockResolvedValue({ id: '3' });
  });

  it('persists supported wizard data before marking the step complete', async () => {
    const { result } = renderHook(() => useCompleteSetupStep(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        stepKey: 'company_profile',
        data: { name: 'Acme', currency: 'EUR', timezone: 'Europe/Paris' },
      });
      await result.current.mutateAsync({
        stepKey: 'work_schedule',
        data: {
          name: 'Standard',
          workingDays: ['monday'],
          startTime: '09:00',
          endTime: '17:00',
          breakMinutes: 30,
          contractedHoursPerWeek: 40,
          dailyOtThresholdMinutes: 480,
          weeklyOtThresholdMinutes: 2400,
        },
      });
      await result.current.mutateAsync({
        stepKey: 'org_structure',
        data: { departments: [{ name: 'Operations' }] },
      });
      await result.current.mutateAsync({
        stepKey: 'locations',
        data: { locations: [{ name: 'HQ', address: '1 Main St' }] },
      });
      await result.current.mutateAsync({
        stepKey: 'payroll',
        data: { payrollPeriod: 'monthly', enableOvertime: true, payDayOfMonth: 28 },
      });
    });

    expect(mocks.put).toHaveBeenCalledWith('/hr/config', {
      companyName: 'Acme',
      defaultCurrency: 'EUR',
      defaultTimezone: 'Europe/Paris',
    });
    expect(mocks.workSchedule.createWorkSchedule).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Standard' })
    );
    expect(mocks.department.createDepartment).toHaveBeenCalledWith({ name: 'Operations' });
    expect(mocks.post).toHaveBeenCalledWith(
      '/hr/locations',
      expect.objectContaining({ name: 'HQ' })
    );
    expect(mocks.put).toHaveBeenCalledWith('/hr/payroll/config', {
      periodType: 'monthly',
      payDayOfMonth: 28,
      overtimeEnabled: true,
    });
    expect(mocks.post).toHaveBeenCalledWith('/hr/setup/steps/company_profile/complete', {});
    expect(mocks.post).toHaveBeenCalledWith('/hr/setup/steps/payroll/complete', {});
  });
});

describe('useGetSetupStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('joins the authoritative HR config onto the legacy status response', async () => {
    const status = {
      steps: [{ stepKey: 'employees', completedAt: '1780000000' }],
      isComplete: false,
    };
    const config = {
      id: '7',
      setupCompletedAt: '1780000000',
      timeTrackingModel: 'both',
    };
    mocks.get.mockImplementation((path: string) => {
      if (path === '/hr/setup/status') return Promise.resolve(status);
      if (path === '/hr/config') return Promise.resolve({ config });
      throw new Error(`Unexpected request: ${path}`);
    });

    const { result } = renderHook(() => useGetSetupStatus(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual({ ...status, config });
    expect(mocks.get).toHaveBeenCalledWith('/hr/setup/status');
    expect(mocks.get).toHaveBeenCalledWith('/hr/config');
  });
});
