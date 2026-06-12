import { unwrap } from '@voxire/api';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/api/posService';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import type { DaySummary } from '../../shared/types/pos';
import { buildLocalDaySummary, normalizeDaySummaryResponse } from './cashDrawerSummary';

export function useDaySummary(date?: string) {
  const [summary, setSummary] = useState<DaySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const openingFloat = useCashDrawerStore((s) => s.openingFloat);
  const movements = useCashDrawerStore((s) => s.movements);
  const expectedCash = useCashDrawerStore((s) => s.expectedCash());

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getDaySummary(date));
      const mapped = normalizeDaySummaryResponse(res, {
        date: date ?? new Date().toISOString().slice(0, 10),
        openingFloat,
        closingCash: expectedCash,
      });
      setSummary(mapped);
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load day summary');
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
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [date]);

  return { summary, loading, error, reload: load };
}
