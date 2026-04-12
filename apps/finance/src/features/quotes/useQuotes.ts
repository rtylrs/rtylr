import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { quoteSchema } from '../../shared/schemas/quotes';

export function useQuotes() {
  return useQuery({
    queryKey: ['quotes'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getQuotes());
      const raw = res ?? [];
      return z.array(quoteSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
