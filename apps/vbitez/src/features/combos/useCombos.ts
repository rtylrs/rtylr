import { unwrap } from '@voxire/api';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { Combo } from '../../shared/types/pos';

export function useCombos() {
  const [combos, setCombos] = useState<Combo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getCombos());
      const raw = res?.combos ?? res?.data ?? [];
      const mapped: Combo[] = (Array.isArray(raw) ? raw : []).map((c: any) => ({
        id: String(c.id ?? Math.random().toString(36).slice(2)),
        name: c.name ?? 'Combo',
        price: Number(c.price ?? 0),
        description: c.description,
        items: Array.isArray(c.items)
          ? c.items.map((it: any) => ({
              itemId: String(it.itemId ?? it.id),
              qty: Number(it.qty ?? 1),
              name: it.name,
            }))
          : [],
      }));
      setCombos(mapped);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load combos');
      setCombos([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return { combos, loading, error, reload: load };
}
