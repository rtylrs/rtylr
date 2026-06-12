import { useEffect, useRef } from 'react';
import { toast } from 'sonner';
import type { RealtimeEvent } from '@voxire/api';
import { posRealtime } from '../services/api/realtime';
import { useUserStore } from '../stores/userStore';

/**
 * Subscribes to stock-out and low-stock realtime events and surfaces
 * actionable toasts to staff. Runs once per session - deduplicates by
 * product ID so the same product only triggers one toast per 5 minutes.
 */
export function useStockAlerts(): void {
  const userId = useUserStore((s) => s.userId);
  const cooldowns = useRef(new Map<string, number>());

  useEffect(() => {
    if (!userId) return;
    const topic = `owner:${userId}`;

    const handler = (e: RealtimeEvent): void => {
      // RealtimeEvent carries extra fields from the server beyond the typed interface
      const raw = e as RealtimeEvent & Record<string, unknown>;

      if (e.type === 'pos:stock:out') {
        const productId = raw['product_id'] as number | undefined;
        const productName = (raw['product_name'] as string | undefined) ?? 'A product';
        const key = `out:${productId}`;
        const last = cooldowns.current.get(key) ?? 0;
        if (Date.now() - last < 5 * 60 * 1000) return;
        cooldowns.current.set(key, Date.now());
        toast.error(`${productName} is out of stock`, {
          description: 'Remove from menu or update inventory.',
          duration: 10000,
        });
      }

      if (e.type === 'pos:stock:low') {
        const productId = raw['product_id'] as number | undefined;
        const productName = (raw['product_name'] as string | undefined) ?? 'A product';
        const currentQty = raw['current_qty'] as number | undefined;
        const threshold = raw['threshold'] as number | undefined;
        const key = `low:${productId}`;
        const last = cooldowns.current.get(key) ?? 0;
        if (Date.now() - last < 5 * 60 * 1000) return;
        cooldowns.current.set(key, Date.now());
        toast.warning(`${productName} running low`, {
          description: `${currentQty ?? '?'} remaining (threshold: ${threshold ?? '?'}).`,
          duration: 8000,
        });
      }
    };

    const unsub = posRealtime.subscribe(topic, handler);
    return () => unsub();
  }, [userId]);
}
