import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { balanceSheetSchema } from '../../shared/schemas/balanceSheet';

export function useBalanceSheet() {
  return useQuery({
    queryKey: ['balance-sheet'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBalanceSheet());
      return balanceSheetSchema.parse(res ?? {});
    },
  });
}
