import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { quoteSchema } from '../../shared/schemas/quotes';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useQuotes() {
  return useQuery({
    queryKey: ['quotes'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/quotes'));
      return z.array(quoteSchema).parse(unwrapList(res, 'quotes'));
    },
  });
}
