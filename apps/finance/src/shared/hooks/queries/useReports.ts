import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { GetPLReportResponse } from '@voxire/voxie/finance/report/v1/get_pl_report';
import { safe } from '../../lib/api';
import { plReportQueryParams } from '../../services/financeApi';

interface ReportRange {
  from: string;
  to: string;
}

export interface BalanceSheetRow {
  label: string;
  amount: number;
  type: string;
}

export interface BalanceSheetResponse {
  assets: BalanceSheetRow[];
  liabilities: BalanceSheetRow[];
  equity: BalanceSheetRow[];
}

interface BalanceSheetApiResponse {
  totalAssets: number;
  totalLiabilities: number;
  totalEquity: number;
  cashBalance: number;
  accountsReceivable: number;
  accountsPayable: number;
}

export interface CashFlowPoint {
  label: string;
  inflow: number;
  outflow: number;
}

export interface CashFlowResponse {
  points: CashFlowPoint[];
}

export const useBalanceSheet = (range: ReportRange) => {
  return useQuery<BalanceSheetResponse>({
    queryKey: ['balance-sheet', range],
    queryFn: async (): Promise<BalanceSheetResponse> => {
      const raw = await unwrap(
        safe.get<BalanceSheetApiResponse>('/finance/balance-sheet', { params: range })
      );
      const res = raw as BalanceSheetApiResponse;
      return {
        assets: [
          { label: 'Cash', amount: res.cashBalance ?? 0, type: 'asset' },
          { label: 'Accounts Receivable', amount: res.accountsReceivable ?? 0, type: 'asset' },
        ],
        liabilities: [
          { label: 'Accounts Payable', amount: res.accountsPayable ?? 0, type: 'liability' },
        ],
        equity: [{ label: 'Equity', amount: res.totalEquity ?? 0, type: 'equity' }],
      };
    },
    enabled: !!range.from && !!range.to,
  });
};

export const useCashFlow = (range: ReportRange) => {
  return useQuery<CashFlowResponse>({
    queryKey: ['cash-flow', range],
    queryFn: async (): Promise<CashFlowResponse> => {
      const res = await unwrap(safe.get<CashFlowResponse>('/finance/cash-flow', { params: range }));
      return res as CashFlowResponse;
    },
    enabled: !!range.from && !!range.to,
  });
};

export const usePL = (range: ReportRange) => {
  return useQuery<GetPLReportResponse>({
    queryKey: ['pl', range],
    queryFn: async (): Promise<GetPLReportResponse> => {
      const res = await unwrap(
        safe.get<GetPLReportResponse>('/api/finance/report/v1/pl', {
          params: plReportQueryParams(range),
        })
      );
      return res as GetPLReportResponse;
    },
    enabled: !!range.from && !!range.to,
  });
};
