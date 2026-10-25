import { useEffect, useRef } from 'react';
import { toast } from 'sonner';
import type { RealtimeEvent } from '@voxire/api';
import { posRealtime } from '../services/api/realtime';
import { useUserStore } from '../stores/userStore';

const COOLDOWN_MS = 5 * 60 * 1000;
// Prune the cooldown map when it grows past this many entries so a
// long-lived terminal tab doesn't accumulate stale keys forever.
const MAX_COOLDOWN_ENTRIES = 500;

function pruneCooldowns(map: Map<string, number>): void {
  if (map.size <= MAX_COOLDOWN_ENTRIES) return;
  const now = Date.now();
  for (const [key, ts] of map) {
    if (now - ts >= COOLDOWN_MS) map.delete(key);
  }
}

/**
 * Subscribes to stock-out and low-stock realtime events and surfaces
 * actionable toasts to staff. Runs once per session - deduplicates by
 * product ID so the same product only triggers one toast per 5 minutes.
 */
export function useStockAlerts(): void {
  const businessId = useUserStore((s) => s.businessId);
  const cooldowns = useRef(new Map<string, number>());

  useEffect(() => {
    if (!businessId) return;
    const topic = `business:${businessId}`;

    const handler = (e: RealtimeEvent): void => {
      // RealtimeEvent carries extra fields from the server beyond the typed interface
      const raw = e as RealtimeEvent & Record<string, unknown>;

      if (e.type === 'pos:stock:out') {
        const productId = raw['product_id'] as number | undefined;
        const productName = (raw['product_name'] as string | undefined) ?? 'A product';
        const key = `out:${productId}`;
        const last = cooldowns.current.get(key) ?? 0;
        if (Date.now() - last < COOLDOWN_MS) return;
        pruneCooldowns(cooldowns.current);
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
        if (Date.now() - last < COOLDOWN_MS) return;
        pruneCooldowns(cooldowns.current);
        cooldowns.current.set(key, Date.now());
        toast.warning(`${productName} running low`, {
          description: `${currentQty ?? '?'} remaining (threshold: ${threshold ?? '?'}).`,
          duration: 8000,
        });
      }
    };

    const unsub = posRealtime.subscribe(topic, handler);
    return () => unsub();
  }, [businessId]);
}
