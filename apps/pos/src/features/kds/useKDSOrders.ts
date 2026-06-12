import { unwrap } from '@voxire/api';
import { useCallback, useEffect, useState } from 'react';
import { posService } from '../../shared/services/api/posService';
import { useUserStore } from '../../shared/stores/userStore';
import { buildKdsStatusUpdate, mapKdsOrders, parsePositiveInteger } from './kdsWorkflow';
import { useKdsStream } from './useKdsStream';

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

  const rawUserId = useUserStore((s) => s.userId);
  const ownerID = rawUserId != null ? parsePositiveInteger(rawUserId) : undefined;

  const load = useCallback(async (): Promise<void> => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getKDSOrders(station));
      const raw = res?.orders ?? res?.data ?? [];
      setOrders(mapKdsOrders(raw));
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load KDS orders');
      setOrders([]);
    } finally {
      setLoading(false);
    }
  }, [station]);

  useEffect(() => {
    void load();
  }, [load]);

  useKdsStream(ownerID, load);

  const setStatus = async (id: string, status: KDSOrder['status']): Promise<void> => {
    const request = buildKdsStatusUpdate(id, status);
    setOrders((prev) => prev.map((o) => (o.id === id ? { ...o, status } : o)));
    try {
      await unwrap(posService.updateKDSItemStatus(request.itemId, request.status));
    } catch {
      // leave local state updated; real sync handled by reload
    }
  };

  return { orders, loading, error, reload: load, setStatus };
}
