import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { ModifierGroup } from '../../shared/types/pos';
import { modifierGroupSchema } from '../../shared/schemas/modifiers';

export function useModifiers() {
  const [groups, setGroups] = useState<ModifierGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await unwrap(posService.getModifiers());
      const raw = res?.groups ?? res?.data ?? [];
      const parsed = z.array(modifierGroupSchema).parse(Array.isArray(raw) ? raw : []);
      const mapped: ModifierGroup[] = parsed.map((g) => ({
        id: g.id,
        name: g.name,
        min: g.min,
        max: g.max,
        modifiers: g.modifiers.map((m) => ({
          id: m.id,
          name: m.name,
          priceDelta: m.priceDelta,
          group: m.group ?? g.name,
          required: m.required,
          defaultSelected: m.defaultSelected,
        })),
      }));
      setGroups(mapped);
    } catch (e: any) {
      setError(e?.response?.data?.message ?? e?.message ?? 'Failed to load modifiers');
      setGroups([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return { groups, loading, error, reload: load };
}
