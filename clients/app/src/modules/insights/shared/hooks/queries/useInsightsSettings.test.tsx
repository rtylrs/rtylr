import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../testUtils/i18nTestSetup';
import { useInsightsSettings, useUpdateInsightsSettings } from './useInsightsSettings';

const api = vi.hoisted(() => ({ getInsightsSettings: vi.fn(), updateInsightsSettings: vi.fn() }));
vi.mock('@voxire/api', () => ({ createOneKitClient: () => api }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
initTestI18n();

const wireSettings = {
  currency: 'EUR',
  defaultPeriod: 'month',
  fiscalYearStart: '04-01',
  timezone: 'Asia/Beirut',
  comparePrevious: true,
  dashboardWidgets: ['revenue'],
  favoriteReports: ['weekly'],
};
const mappedSettings = {
  defaultCurrency: 'EUR',
  defaultPeriod: 'month',
  fiscalYearStart: '04-01',
  timezone: 'Asia/Beirut',
  comparePrevious: true,
  dashboardWidgets: ['revenue'],
  favoriteReports: ['weekly'],
};
function setup() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return { client, wrapper };
}
beforeEach(() => {
  vi.clearAllMocks();
  api.getInsightsSettings.mockResolvedValue({ settings: wireSettings });
});
describe('Insights settings persistence', () => {
  it('loads the stored fiscal year and timezone', async () => {
    const { wrapper } = setup();
    const { result } = renderHook(() => useInsightsSettings(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mappedSettings);
  });
  it('saves changed values while preserving the other settings', async () => {
    const { client, wrapper } = setup();
    client.setQueryData(['insights-settings'], mappedSettings);
    api.updateInsightsSettings.mockResolvedValue({
      settings: { ...wireSettings, timezone: 'Europe/Paris', fiscalYearStart: '07-01' },
    });
    const { result } = renderHook(() => useUpdateInsightsSettings(), { wrapper });
    await act(() =>
      result.current.mutateAsync({ timezone: 'Europe/Paris', fiscalYearStart: '07-01' })
    );
    expect(api.updateInsightsSettings).toHaveBeenCalledWith({
      settings: { ...wireSettings, timezone: 'Europe/Paris', fiscalYearStart: '07-01' },
    });
    expect(client.getQueryData(['insights-settings'])).toEqual({
      ...mappedSettings,
      timezone: 'Europe/Paris',
      fiscalYearStart: '07-01',
    });
  });
  it('leaves stored values intact when saving fails', async () => {
    const { client, wrapper } = setup();
    client.setQueryData(['insights-settings'], mappedSettings);
    api.updateInsightsSettings.mockRejectedValue(new Error('Failed to save'));
    const { result } = renderHook(() => useUpdateInsightsSettings(), { wrapper });
    await act(async () => {
      await expect(result.current.mutateAsync({ timezone: 'Europe/Paris' })).rejects.toThrow(
        'Failed to save'
      );
    });
    expect(client.getQueryData(['insights-settings'])).toEqual(mappedSettings);
  });
});
