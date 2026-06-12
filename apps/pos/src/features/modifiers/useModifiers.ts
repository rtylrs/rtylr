import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { posService } from '../../shared/services/api/posService';
import type { ModifierGroup } from '../../shared/types/pos';
import { modifierGroupSchema } from '../../shared/utils/schemas/modifiers';

export function useModifiers() {
  const query = useQuery({
    queryKey: ['pos-modifiers'],
    queryFn: async (): Promise<ModifierGroup[]> => {
      const res = await unwrap(posService.getModifiers());
      const raw = res?.groups ?? res?.data ?? [];
      const parsed = z.array(modifierGroupSchema).parse(Array.isArray(raw) ? raw : []);
      return parsed.map((g) => ({
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
    },
    staleTime: 5 * 60 * 1000,
  });

  return {
    groups: query.data ?? [],
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: query.refetch,
  };
}
