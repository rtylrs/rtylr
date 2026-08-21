import { createOneKitClient } from '@voxire/api';
import { DaySummaryServiceClient } from '@voxire/voxie/pos/day_summary/v1/client';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import type { DaySummary } from '../../shared/types/pos';
import { buildLocalDaySummary, normalizeDaySummaryResponse } from './cashDrawerSummary';

const daySummaryClient = createOneKitClient(DaySummaryServiceClient, 'pos');

export function useDaySummary(date?: string) {
  const [summary, setSummary] = useState<DaySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const openingFloat = useCashDrawerStore((s) => s.openingFloat);
  const movements = useCashDrawerStore((s) => s.movements);
  const expectedCashFn = useCashDrawerStore((s) => s.expectedCash);
  // Compute from selected slices instead of running a reduce inside the
  // store selector on every notification.
  const expectedCash = useMemo(() => expectedCashFn(), [expectedCashFn, movements, openingFloat]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await daySummaryClient.daySummary({ date: date ?? '' });
      const mapped = normalizeDaySummaryResponse(res, {
        date: date ?? new Date().toISOString().slice(0, 10),
        openingFloat,
        closingCash: expectedCash,
      });
      setSummary(mapped);
    } catch (e) {
      setError(e instanceof Error && e.message ? e.message : 'Failed to load day summary');
      setSummary(
        buildLocalDaySummary({
          date: date ?? new Date().toISOString().slice(0, 10),
          openingFloat,
          movements,
          closingCash: expectedCash,
        })
      );
    } finally {
      setLoading(false);
    }
  }, [date, openingFloat, movements, expectedCash]);

  useEffect(() => {
    void load();
  }, [load]);

  return { summary, loading, error, reload: load };
}
