import { unwrap } from '@voxire/api';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import type { DaySummary } from '../../shared/types/pos';
import { daySummarySchema } from '../../shared/schemas/daySummary';

export function useDaySummary(date?: string) {
  const [summary, setSummary] = useState<DaySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const openingFloat = useCashDrawerStore((s) => s.openingFloat);
  const expectedCash = useCashDrawerStore((s) => s.expectedCash());

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getDaySummary(date));
      const raw = res?.data ?? {};
      const parsed = daySummarySchema.parse({
        ...raw,
        date: raw.date ?? date,
        openingFloat: raw.openingFloat ?? openingFloat,
        closingCash: raw.closingCash ?? expectedCash,
      });
      const mapped: DaySummary = parsed;
      setSummary(mapped);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load day summary');
      setSummary({
        date: date ?? new Date().toISOString().slice(0, 10),
        grossSales: 0,
        netSales: 0,
        taxCollected: 0,
        tips: 0,
        refunds: 0,
        voids: 0,
        orderCount: 0,
        averageTicket: 0,
        cashDrops: 0,
        openingFloat,
        closingCash: expectedCash,
      });
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
