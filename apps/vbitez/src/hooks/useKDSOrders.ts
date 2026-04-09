import { useEffect, useState } from 'react';
import { posService } from '../services/posService';

export interface KDSOrder {
  id: string;
  number: string;
  station: string;
  items: { name: string; qty: number; course?: string; notes?: string }[];
  status: 'new' | 'cooking' | 'ready' | 'served';
  createdAt: string;
  elapsedSeconds?: number;
}

export function useKDSOrders(station?: string) {
  const [orders, setOrders] = useState<KDSOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await posService.getKDSOrders(station);
      const raw = res?.data?.orders ?? res?.data ?? [];
      const mapped: KDSOrder[] = (Array.isArray(raw) ? raw : []).map((o: any) => ({
        id: String(o.id ?? Math.random().toString(36).slice(2)),
        number: String(o.number ?? o.id ?? ''),
        station: o.station ?? 'Kitchen',
        items: Array.isArray(o.items)
          ? o.items.map((it: any) => ({
              name: it.name ?? 'Item',
              qty: Number(it.qty ?? 1),
              course: it.course,
              notes: it.notes,
            }))
          : [],
        status: (o.status ?? 'new') as KDSOrder['status'],
        createdAt: o.createdAt ?? new Date().toISOString(),
      }));
      setOrders(mapped);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load KDS orders');
      setOrders([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const t = window.setInterval(load, 15000);
    return () => window.clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [station]);

  const setStatus = async (id: string, status: KDSOrder['status']) => {
    setOrders((prev) => prev.map((o) => (o.id === id ? { ...o, status } : o)));
    try {
      await posService.updateKDSStatus(id, { status });
    } catch {
      // leave local state updated; real sync handled by reload
    }
  };

  return { orders, loading, error, reload: load, setStatus };
}
