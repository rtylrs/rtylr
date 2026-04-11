import { useEffect, useState } from 'react';
import { posService } from '../../shared/services/posService';
import type { ModifierGroup } from '../../shared/types/pos';

export function useModifiers() {
  const [groups, setGroups] = useState<ModifierGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await posService.getModifiers();
      const raw = res?.data?.groups ?? res?.data ?? [];
      const mapped: ModifierGroup[] = (Array.isArray(raw) ? raw : []).map((g: any) => ({
        id: String(g.id ?? Math.random().toString(36).slice(2)),
        name: g.name ?? 'Modifiers',
        min: Number(g.min ?? 0),
        max: Number(g.max ?? 1),
        modifiers: (g.modifiers ?? []).map((m: any) => ({
          id: String(m.id ?? Math.random().toString(36).slice(2)),
          name: m.name ?? 'Option',
          priceDelta: Number(m.priceDelta ?? 0),
          group: g.name,
          required: !!m.required,
          defaultSelected: !!m.default,
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
