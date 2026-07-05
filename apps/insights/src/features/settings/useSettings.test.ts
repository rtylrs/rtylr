import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getInsightsSettings: vi.fn(),
    updateInsightsSettings: vi.fn(),
    getScheduledReports: vi.fn(),
    createScheduledReport: vi.fn(),
    deleteScheduledReport: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import {
  useInsightsSettings,
  useUpdateInsightsSettings,
  useScheduledReports,
  useCreateScheduledReport,
  useDeleteScheduledReport,
} from './useSettings';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useInsightsSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches settings on mount', async () => {
    const data = { defaultCurrency: 'USD', fiscalYearStart: '01-01', timezone: 'UTC' };
    (insightsService.getInsightsSettings as any).mockResolvedValue(data);
    const { result } = renderHook(() => useInsightsSettings(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (insightsService.getInsightsSettings as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useInsightsSettings(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useUpdateInsightsSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateInsightsSettings with provided data', async () => {
    (insightsService.updateInsightsSettings as any).mockResolvedValue({ defaultCurrency: 'EUR' });
    const { result } = renderHook(() => useUpdateInsightsSettings(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ defaultCurrency: 'EUR' });
    });
    expect(insightsService.updateInsightsSettings).toHaveBeenCalledWith({ defaultCurrency: 'EUR' });
  });
});

describe('useScheduledReports', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches scheduled reports', async () => {
    const response = {
      reports: [{ id: '1', name: 'Weekly Sales', frequency: 'weekly', enabled: true }],
    };
    (insightsService.getScheduledReports as any).mockResolvedValue(response);
    const { result } = renderHook(() => useScheduledReports(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
    expect(result.current.data).toHaveLength(1);
  });
});

describe('useCreateScheduledReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createScheduledReport with provided payload', async () => {
    (insightsService.createScheduledReport as any).mockResolvedValue({ id: '2' });
    const { result } = renderHook(() => useCreateScheduledReport(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({
        name: 'Monthly Revenue',
        templateId: 'revenue',
        frequency: 'monthly',
        recipients: ['admin@example.com'],
        enabled: true,
      });
    });
    expect(insightsService.createScheduledReport).toHaveBeenCalled();
  });
});

describe('useDeleteScheduledReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteScheduledReport with the given id', async () => {
    (insightsService.deleteScheduledReport as any).mockResolvedValue({});
    const { result } = renderHook(() => useDeleteScheduledReport(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync('report-123');
    });
    expect(insightsService.deleteScheduledReport).toHaveBeenCalledWith('report-123');
  });
});
