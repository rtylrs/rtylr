import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { billSchema } from '../../shared/schemas/bills';

export function useBills() {
  return useQuery({
    queryKey: ['bills'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBills());
      const raw = res ?? [];
      return z.array(billSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
