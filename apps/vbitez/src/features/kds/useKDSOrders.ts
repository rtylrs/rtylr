import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import { kdsOrderSchema } from '../../shared/schemas/kdsOrders';

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
      const res = await unwrap(posService.getKDSOrders(station));
      const raw = res?.orders ?? res?.data ?? [];
      const parsed = z.array(kdsOrderSchema).parse(Array.isArray(raw) ? raw : []);
      const mapped: KDSOrder[] = parsed.map((o) => ({
        id: o.id,
        number: o.number ?? o.id,
        station: o.station,
        items: o.items.map((it) => ({
          name: it.name,
          qty: it.qty,
          course: it.course ?? undefined,
          notes: it.notes ?? undefined,
        })),
        status: o.status,
        createdAt: o.createdAt,
        elapsedSeconds: o.elapsedSeconds,
      }));
      setOrders(mapped);
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load KDS orders');
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
      await unwrap(posService.updateKDSItemStatus(Number(id), status));
    } catch {
      // leave local state updated; real sync handled by reload
    }
  };

  return { orders, loading, error, reload: load, setStatus };
}
