import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  listChecklistTemplates: vi.fn(),
  createChecklistTemplate: vi.fn(),
  listEmployeeChecklists: vi.fn(),
  listStaff: vi.fn(),
  completeItem: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    listChecklistTemplates: mocks.listChecklistTemplates,
    createChecklistTemplate: mocks.createChecklistTemplate,
    listEmployeeChecklists: mocks.listEmployeeChecklists,
    listStaff: mocks.listStaff,
  }),
  unwrap: (value: unknown) => value,
}));
vi.mock('../../lib/clients', () => ({
  hrTransport: { post: mocks.completeItem },
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import {
  useGetEmployeeChecklist,
  useListChecklistTemplates,
  useListEmployeeChecklistSummaries,
} from './useOnboarding';

const wrapper = ({ children }: { children: React.ReactNode }): React.JSX.Element => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return React.createElement(QueryClientProvider, { client }, children);
};

describe('onboarding contract hooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.listChecklistTemplates.mockResolvedValue({ templates: [] });
    mocks.listEmployeeChecklists.mockResolvedValue({ checklists: [] });
    mocks.listStaff.mockResolvedValue({ staff: [] });
  });

  it('reads checklist templates through the generated versioned client', async () => {
    mocks.listChecklistTemplates.mockResolvedValue({ templates: [{ id: '7', name: 'New hire' }] });
    const { result } = renderHook(() => useListChecklistTemplates(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.listChecklistTemplates).toHaveBeenCalledWith({});
    expect(result.current.data?.templates?.[0]?.name).toBe('New hire');
  });

  it('uses the generated employee-checklist list contract', async () => {
    const { result } = renderHook(() => useGetEmployeeChecklist(42), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.listEmployeeChecklists).toHaveBeenCalledWith({
      employeeId: '42',
      limit: 100,
      offset: 0,
    });
  });

  it('derives management summaries from supported employee checklist data', async () => {
    mocks.listStaff.mockResolvedValue({ staff: [{ id: '42' }] });
    mocks.listEmployeeChecklists.mockResolvedValue({
      checklists: [
        {
          id: '9',
          employeeId: '42',
          templateId: '3',
          triggeredAt: '1700000000',
          createdAt: '1700000000',
          items: [
            { status: 'complete', dueDate: '2020-01-01' },
            { status: 'incomplete', dueDate: '2020-01-02' },
          ],
        },
      ],
    });

    const { result } = renderHook(
      () => useListEmployeeChecklistSummaries({ limit: 20, offset: 0 }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.summaries).toEqual([
      expect.objectContaining({
        employeeId: 42,
        checklistId: 9,
        totalItems: 2,
        completedItems: 1,
        completionPercentage: 50,
        overdueItemsCount: 1,
      }),
    ]);
    expect(mocks.listEmployeeChecklists).toHaveBeenCalledWith({
      employeeId: '42',
      limit: 100,
      offset: 0,
    });
  });
});
