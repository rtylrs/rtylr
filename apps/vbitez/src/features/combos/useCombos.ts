import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { Combo } from '../../shared/types/pos';
import { comboSchema } from '../../shared/schemas/combos';

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
      const parsed = z.array(comboSchema).parse(Array.isArray(raw) ? raw : []);
      const mapped: Combo[] = parsed.map((c) => ({
        id: c.id,
        name: c.name,
        price: c.price,
        description: c.description ?? undefined,
        items: c.items,
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
