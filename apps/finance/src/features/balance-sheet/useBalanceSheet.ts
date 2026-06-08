import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { balanceSheetSchema } from '../../shared/schemas/balanceSheet';

export function useBalanceSheet() {
  const asOf = new Date().toISOString().slice(0, 10);
  return useQuery({
    queryKey: ['balance-sheet', asOf],
    queryFn: async () => {
      const res = await unwrap(
        safe.get<unknown>('/finance/balance-sheet', { params: { as_of: asOf } })
      );
      return balanceSheetSchema.parse(res ?? {});
    },
  });
}
