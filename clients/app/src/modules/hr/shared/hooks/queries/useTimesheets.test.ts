import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}));

vi.mock('../../lib/clients', () => ({ hrTransport: mocks }));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useSubmitTimesheet } from './useTimesheets';

const wrapperFor =
  (client: QueryClient): React.FC<{ children: React.ReactNode }> =>
  ({ children }: { children: React.ReactNode }): React.JSX.Element =>
    React.createElement(QueryClientProvider, { client }, children);

describe('useSubmitTimesheet', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // TimesheetDetail reads the singular ['timesheet', id] key, which the plural
  // invalidation does not match. Without it the detail stayed cached as draft
  // and offered Submit again for a sheet the server had already submitted.
  it('invalidates the submitted timesheet detail, not just the list', async () => {
    mocks.post.mockResolvedValue({ success: true });
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    client.setQueryData(['hr', 'timesheet', 7], { id: 7, status: 'draft' });

    const { result } = renderHook(() => useSubmitTimesheet(), { wrapper: wrapperFor(client) });
    await act(() => result.current.mutateAsync({ timesheetId: 7 }));

    expect(mocks.post).toHaveBeenCalledWith('/hr/timesheets/7/submit', {});
    expect(client.getQueryState(['hr', 'timesheet', 7])?.isInvalidated).toBe(true);
  });
});
