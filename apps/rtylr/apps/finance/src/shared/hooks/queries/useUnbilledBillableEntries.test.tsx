// @vitest-environment happy-dom
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockClient = vi.hoisted(() => ({
  listTimeEntries: vi.fn(),
  markTimeEntriesInvoiced: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
}));
vi.mock('@voxire/voxie/hr/time_entry/v1/client', () => ({
  TimeEntryServiceClient: {},
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useUnbilledBillableEntries } from './useUnbilledBillableEntries';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

const entries = Array.from({ length: 101 }, (_, index) => ({
  id: `entry-${index}`,
  date: '2026-09-01',
  description: `Entry ${index}`,
  durationMinutes: 30,
  projectId: 'project-1',
}));

describe('useUnbilledBillableEntries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('follows the response total past the server page-size cap', async () => {
    mockClient.listTimeEntries.mockImplementation(async ({ offset }: { offset: number }) => ({
      entries: entries.slice(offset, offset + 100),
      total: entries.length,
      limit: 100,
      offset,
    }));

    const { result } = renderHook(() => useUnbilledBillableEntries('2026-06-01', '2026-09-01'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(101);
    expect(result.current.data?.[100]?.id).toBe('entry-100');
    expect(mockClient.listTimeEntries).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ offset: 0, limit: 200 })
    );
    expect(mockClient.listTimeEntries).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ offset: 100, limit: 200 })
    );
  });
});
