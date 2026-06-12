import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { posService } from '../../shared/services/api/posService';
import type { Combo } from '../../shared/types/pos';
import { comboSchema } from '../../shared/utils/schemas/combos';

export function useCombos() {
  const query = useQuery({
    queryKey: ['pos-combos'],
    queryFn: async (): Promise<Combo[]> => {
      const res = await unwrap(posService.getCombos());
      const raw = res?.combos ?? res?.data ?? [];
      const parsed = z.array(comboSchema).parse(Array.isArray(raw) ? raw : []);
      return parsed.map((c) => ({
        id: c.id,
        name: c.name,
        price: c.price,
        description: c.description ?? undefined,
        items: c.items,
      }));
    },
    staleTime: 5 * 60 * 1000,
  });

  return {
    combos: query.data ?? [],
    loading: query.isLoading,
    error: query.error
      ? query.error instanceof Error
        ? query.error.message
        : String(query.error)
      : null,
    reload: query.refetch,
  };
}
