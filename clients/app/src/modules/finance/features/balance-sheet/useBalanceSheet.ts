import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { StatementServiceClient } from '@voxire/voxie/finance/statement/v1/client';
import { balanceSheetSchema } from '../../shared/schemas/balanceSheet';

const statementClient = createOneKitClient(StatementServiceClient, 'finance');

/** The balance sheet on a date (YYYY-MM-DD). */
export function useBalanceSheet(asOf: string) {
  return useQuery({
    queryKey: ['finance', 'balance-sheet', asOf],
    queryFn: async () => balanceSheetSchema.parse(await statementClient.getBalanceSheet({ asOf })),
  });
}
