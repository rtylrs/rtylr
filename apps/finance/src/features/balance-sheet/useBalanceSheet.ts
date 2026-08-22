import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { StatementServiceClient } from '@voxire/voxie/finance/statement/v1/client';
import { balanceSheetSchema } from '../../shared/schemas/balanceSheet';

const statementClient = createOneKitClient(StatementServiceClient, 'finance');

// finance.statement.v1.GetBalanceSheet has no as-of-date param - it's always
// the current point-in-time balance sheet, so no date can be passed.
export function useBalanceSheet() {
  return useQuery({
    queryKey: ['balance-sheet'],
    queryFn: async () => {
      const res = await statementClient.getBalanceSheet({});
      return balanceSheetSchema.parse({
        assets: [
          { label: 'Cash', amount: res.cashBalance, type: 'asset' },
          { label: 'Accounts Receivable', amount: res.accountsReceivable, type: 'asset' },
        ],
        liabilities: [
          { label: 'Accounts Payable', amount: res.accountsPayable, type: 'liability' },
        ],
        equity: [{ label: 'Equity', amount: res.totalEquity, type: 'equity' }],
        totalAssets: res.totalAssets,
        totalLiabilities: res.totalLiabilities,
        totalEquity: res.totalEquity,
      });
    },
  });
}
