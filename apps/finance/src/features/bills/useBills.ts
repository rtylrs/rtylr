import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { billSchema } from '../../shared/schemas/bills';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useBills() {
  return useQuery({
    queryKey: ['bills'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBills());
      return z.array(billSchema).parse(unwrapList(res, 'bills'));
    },
  });
}
