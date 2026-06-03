import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getReportTemplates: vi.fn(),
    generateReport: vi.fn(),
    exportReport: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import { useReportTemplates, useGenerateReport, useExportReport } from './useReports';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useReportTemplates', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches report templates on mount', async () => {
    const templates = [{ id: '1', name: 'Sales Summary', description: 'Sales overview' }];
    (insightsService.getReportTemplates as any).mockResolvedValue(templates);
    const { result } = renderHook(() => useReportTemplates(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
  });

  it('unwraps the "templates" key when present', async () => {
    const data = {
      templates: [
        { id: 'revenue-monthly', name: 'Revenue Report', description: 'Monthly revenue' },
      ],
    };
    (insightsService.getReportTemplates as any).mockResolvedValue(data);
    const { result } = renderHook(() => useReportTemplates(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
  });

  it('starts in loading state', () => {
    (insightsService.getReportTemplates as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useReportTemplates(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useGenerateReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls generateReport with templateId and params', async () => {
    (insightsService.generateReport as any).mockResolvedValue({ data: 'ok' });
    const { result } = renderHook(() => useGenerateReport(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ templateId: 'sales', params: { period: 'monthly' } });
    });
    expect(insightsService.generateReport).toHaveBeenCalledWith('sales', { period: 'monthly' });
  });

  it('calls generateReport without params when omitted', async () => {
    (insightsService.generateReport as any).mockResolvedValue({});
    const { result } = renderHook(() => useGenerateReport(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ templateId: 'revenue' });
    });
    expect(insightsService.generateReport).toHaveBeenCalledWith('revenue', undefined);
  });
});

describe('useExportReport', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls exportReport with templateId, format, and params', async () => {
    (insightsService.exportReport as any).mockResolvedValue({ url: 'blob://...' });
    const { result } = renderHook(() => useExportReport(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ templateId: 'sales', format: 'csv' });
    });
    expect(insightsService.exportReport).toHaveBeenCalledWith('sales', 'csv', undefined);
  });
});
