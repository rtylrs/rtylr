import { createOneKitClient } from '@voxire/api';
import { useCallback, useEffect, useRef, useState } from 'react';
import { KDSServiceClient } from '@voxire/voxie/pos/kds/v1/client';
import { useUserStore } from '../../shared/stores/userStore';
import { useKdsStream } from './useKdsStream';

export interface KDSOrder {
  id: string;
  number: string;
  station: string;
  items: { id: string; orderId: string; name: string; qty: number; course?: string; notes?: string }[];
  status: 'new' | 'cooking' | 'ready' | 'served';
  createdAt: string;
  elapsedSeconds?: number;
}

const kdsClient = createOneKitClient(KDSServiceClient, 'pos');
const kdsStatuses = new Set<KDSOrder['status']>(['new', 'cooking', 'ready', 'served']);

function parsePositiveInteger(value: unknown): number | undefined {
  const numeric = Number(value);
  return Number.isInteger(numeric) && numeric > 0 ? numeric : undefined;
}

function kdsStatus(value: string): KDSOrder['status'] {
  const lower = (value ?? '').toLowerCase();
  if (lower === 'cooking' || lower === 'preparing') return 'cooking';
  if (lower === 'ready') return 'ready';
  if (lower === 'served') return 'served';
  return kdsStatuses.has(lower as KDSOrder['status']) ? (lower as KDSOrder['status']) : 'new';
}

function toServerItemStatus(status: KDSOrder['status']): string {
  if (status === 'cooking') return 'preparing';
  if (status === 'new') return 'pending';
  return status;
}

export function useKDSOrders(station?: string) {
  const [orders, setOrders] = useState<KDSOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const ordersRef = useRef<KDSOrder[]>([]);

  useEffect(() => {
    ordersRef.current = orders;
  }, [orders]);

  const rawUserId = useUserStore((s) => s.userId);
  const ownerID = rawUserId != null ? parsePositiveInteger(rawUserId) : undefined;

  const load = useCallback(async (options?: { silent?: boolean }): Promise<void> => {
    // Silent reloads (realtime pings) keep the current tickets on screen
    // instead of flashing the skeleton grid on a busy kitchen line.
    if (!options?.silent) setLoading(true);
    setError(null);
    try {
      const res = await kdsClient.listKDSOrders({});
      const nextOrders = (res.orders ?? [])
        .map((order) => ({
          id: order.id ?? '',
          number: order.orderNumber || order.id || '',
          station: order.type || 'Kitchen',
          items: (order.items ?? []).map((item) => ({
            id: item.id ?? '',
            orderId: item.orderId || order.id || '',
            name: item.name || 'Item',
            qty: item.quantity || 1,
            notes: item.notes || undefined,
          })),
          status: kdsStatus(order.status ?? ''),
          createdAt: order.createdAt || new Date().toISOString(),
        }))
        .filter((order) => !station || order.station.toLowerCase() === station.toLowerCase());
      setOrders(nextOrders);
    } catch (e) {
      const err = e as { response?: { data?: { message?: string } }; message?: string };
      setError(err?.response?.data?.message ?? err?.message ?? 'Failed to load KDS orders');
      setOrders([]);
    } finally {
      if (!options?.silent) setLoading(false);
    }
  }, [station]);

  useEffect(() => {
    void load();
  }, [load]);

  useKdsStream(ownerID, () => void load({ silent: true }));

  const setStatus = useCallback(
    async (id: string, status: KDSOrder['status']): Promise<void> => {
      // Read the latest committed orders from the ref, not a render closure,
      // so the callback identity stays stable across 1s tick re-renders.
      const order = ordersRef.current.find((item) => item.id === id);
      const itemUpdates = order?.items.filter((item) => parsePositiveInteger(item.id));
      if (!order || !itemUpdates || itemUpdates.length === 0) {
        return;
      }
      setOrders((prev) => prev.map((o) => (o.id === id ? { ...o, status } : o)));
      try {
        await Promise.all(
          itemUpdates.map((item) =>
            kdsClient.updateKDSItemStatus({
              orderId: item.orderId || order.id,
              itemId: item.id,
              body: { status: toServerItemStatus(status) },
            }),
          ),
        );
      } catch {
        // Optimistic update failed: revert and surface the failure so staff
        // are not left looking at a ticket state the server never accepted.
        setOrders((prev) => prev.map((o) => (o.id === id ? { ...o, status: order.status } : o)));
        setError('Could not update ticket status. Please try again.');
      }
    },
    [],
  );

  return { orders, loading, error, reload: load, setStatus };
}
