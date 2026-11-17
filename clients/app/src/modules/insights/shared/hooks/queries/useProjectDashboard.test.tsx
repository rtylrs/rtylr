import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  requestErrorStatus,
  useDashboardProjects,
  useProjectDashboard,
} from './useInsightsDashboard';

const api = vi.hoisted(() => ({ getProjectDashboard: vi.fn(), listProjects: vi.fn() }));
vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return { ...actual, createOneKitClient: () => api };
});

function wrapper({ children }: { children: React.ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('useProjectDashboard', () => {
  it('maps the wire response, int64 strings included', async () => {
    api.getProjectDashboard.mockResolvedValue({
      project: { projectId: '8', projectName: 'Website', status: 'active', visibility: 'public' },
      progress: { totalTasks: '8', completedTasks: '2', percent: 25 },
      overdueTaskCount: '3',
      overdueTasks: [
        {
          taskId: '51',
          title: 'Fix footer',
          dueDate: '2026-10-01',
          assigneeId: '7',
          assigneeName: 'Abed Amouneh',
        },
        { taskId: '52', title: 'Write copy', dueDate: '2026-10-02' },
      ],
      hours: {
        thisMonth: { value: '2400', deltaPct: 50 },
        priorMinutes: '1600',
        allTimeMinutes: '9000',
      },
      invoices: {
        invoiceCount: '4',
        invoiced: '200000',
        paid: '120000',
        outstanding: '80000',
        overdue: '30000',
        overdueCount: '1',
      },
    });
    const { result } = renderHook(() => useProjectDashboard('8'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(api.getProjectDashboard).toHaveBeenCalledWith({ projectId: '8' });
    expect(result.current.data).toEqual({
      project: { id: '8', name: 'Website', status: 'active', visibility: 'public' },
      totalTasks: 8,
      completedTasks: 2,
      progressPct: 25,
      overdueTaskCount: 3,
      overdueTasks: [
        {
          id: '51',
          title: 'Fix footer',
          dueDate: '2026-10-01',
          assigneeId: '7',
          assigneeName: 'Abed Amouneh',
        },
        { id: '52', title: 'Write copy', dueDate: '2026-10-02', assigneeId: '', assigneeName: '' },
      ],
      hoursThisMonth: { value: 2400, deltaPct: 50 },
      priorMinutes: 1600,
      allTimeMinutes: 9000,
      invoices: {
        invoiceCount: 4,
        invoiced: 200000,
        paid: 120000,
        outstanding: 80000,
        overdue: 30000,
        overdueCount: 1,
      },
    });
  });

  // Zero numbers are omitted on the wire, and invoices are absent without
  // finance access: the hook defaults the numbers and keeps invoices absent.
  it('defaults omitted zeros and keeps a missing invoice section missing', async () => {
    api.getProjectDashboard.mockResolvedValue({
      project: { projectId: '8', projectName: 'Website', visibility: 'private' },
      progress: {},
      hours: { thisMonth: {} },
    });
    const { result } = renderHook(() => useProjectDashboard('8'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toMatchObject({
      totalTasks: 0,
      completedTasks: 0,
      progressPct: 0,
      overdueTaskCount: 0,
      overdueTasks: [],
      hoursThisMonth: { value: 0, deltaPct: 0 },
      priorMinutes: 0,
      allTimeMinutes: 0,
    });
    expect(result.current.data?.invoices).toBeUndefined();
  });

  it('does not ask without a project', () => {
    renderHook(() => useProjectDashboard(''), { wrapper });
    expect(api.getProjectDashboard).not.toHaveBeenCalled();
  });
});

describe('useDashboardProjects', () => {
  it('lists the projects with their ids as strings', async () => {
    api.listProjects.mockResolvedValue({
      projects: [
        { projectId: '8', projectName: 'Website', status: 'active', visibility: 'public' },
        { projectName: 'No id' },
      ],
    });
    const { result } = renderHook(() => useDashboardProjects(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([
      { id: '8', name: 'Website', status: 'active', visibility: 'public' },
    ]);
  });
});

describe('requestErrorStatus', () => {
  it('reads the status wherever the client put it', () => {
    expect(requestErrorStatus({ statusCode: 404 })).toBe(404);
    expect(requestErrorStatus({ status: 403 })).toBe(403);
    expect(requestErrorStatus({ response: { status: 500 } })).toBe(500);
    expect(requestErrorStatus(new Error('offline'))).toBeUndefined();
    expect(requestErrorStatus(null)).toBeUndefined();
  });
});
