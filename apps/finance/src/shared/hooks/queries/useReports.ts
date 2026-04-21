import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { GetPLReportResponse } from '@voxire/contracts/finance/v1/reports';
import { safe } from '../../lib/api';

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
      const res = await unwrap(
        safe.get<BalanceSheetResponse>('/finance/balance-sheet', { params: range })
      );
      return res as BalanceSheetResponse;
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
      const res = await unwrap(safe.get<GetPLReportResponse>('/finance/pl', { params: range }));
      return res as GetPLReportResponse;
    },
    enabled: !!range.from && !!range.to,
  });
};
