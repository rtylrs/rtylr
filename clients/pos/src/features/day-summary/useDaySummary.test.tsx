import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';

const { daySummaryMock } = vi.hoisted(() => ({
  daySummaryMock: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    daySummary: daySummaryMock,
  }),
}));

import { useDaySummary } from './useDaySummary';

describe('useDaySummary', () => {
  beforeEach(() => {
    daySummaryMock.mockReset();
    useCashDrawerStore.setState({
      openingFloat: 50,
      movements: [],
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('requests today in YYYY-MM-DD when no date argument is supplied', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(new Date('2026-11-21T15:30:00.000Z'));

    daySummaryMock.mockResolvedValue({
      totalSales: 120,
      totalTax: 12,
      totalTips: 5,
      totalRefunds: 10,
      totalOrders: 3,
      paymentBreakdown: {},
    });

    const { result } = renderHook(() => useDaySummary());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(daySummaryMock).toHaveBeenCalledWith({ date: '2026-11-21' });
    expect(result.current.error).toBeNull();
    expect(result.current.summary?.date).toBe('2026-11-21');
    expect(result.current.summary?.grossSales).toBe(120);
  });

  it('passes an explicit date through unchanged', async () => {
    daySummaryMock.mockResolvedValue({
      totalSales: 40,
      totalTax: 4,
      totalTips: 0,
      totalRefunds: 0,
      totalOrders: 1,
      paymentBreakdown: {},
    });

    const { result } = renderHook(() => useDaySummary('2026-11-19'));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(daySummaryMock).toHaveBeenCalledWith({ date: '2026-11-19' });
    expect(result.current.summary?.date).toBe('2026-11-19');
  });
});
