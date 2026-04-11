import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import { useCashDrawerStore } from '../../shared/stores/cashDrawerStore';
import type { DaySummary } from '../../shared/types/pos';

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
      const res = await posService.getDaySummary(date);
      const raw = res?.data ?? {};
      const mapped: DaySummary = {
        date: raw.date ?? date ?? new Date().toISOString().slice(0, 10),
        grossSales: Number(raw.grossSales ?? 0),
        netSales: Number(raw.netSales ?? 0),
        taxCollected: Number(raw.taxCollected ?? 0),
        tips: Number(raw.tips ?? 0),
        refunds: Number(raw.refunds ?? 0),
        voids: Number(raw.voids ?? 0),
        orderCount: Number(raw.orderCount ?? 0),
        averageTicket: Number(raw.averageTicket ?? 0),
        cashDrops: Number(raw.cashDrops ?? 0),
        openingFloat: Number(raw.openingFloat ?? openingFloat),
        closingCash: Number(raw.closingCash ?? expectedCash),
      };
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
