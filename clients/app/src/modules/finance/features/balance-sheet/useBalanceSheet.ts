import { createOneKitClient } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { StatementServiceClient } from '@voxire/voxie/finance/statement/v1/client';
import { balanceSheetSchema } from '../../shared/schemas/balanceSheet';

const statementClient = createOneKitClient(StatementServiceClient, 'finance');

// finance.statement.v1.GetBalanceSheet has no as-of-date param - it's always
// the current point-in-time balance sheet.
export function useBalanceSheet() {
  const { t } = useTranslation('financeCommon');
  return useQuery({
    queryKey: ['finance', 'balance-sheet'],
    queryFn: async () => {
      const res = await statementClient.getBalanceSheet({});
      return balanceSheetSchema.parse({
        assets: [
          { label: t('finance.balanceSheet.rows.cash'), amount: res.cashBalance, type: 'asset' },
          {
            label: t('finance.balanceSheet.rows.receivable'),
            amount: res.accountsReceivable,
            type: 'asset',
          },
        ],
        liabilities: [
          {
            label: t('finance.balanceSheet.rows.payable'),
            amount: res.accountsPayable,
            type: 'liability',
          },
        ],
        equity: [
          { label: t('finance.balanceSheet.rows.equity'), amount: res.totalEquity, type: 'equity' },
        ],
        totalAssets: res.totalAssets,
        totalLiabilities: res.totalLiabilities,
        totalEquity: res.totalEquity,
      });
    },
  });
}
