import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
import type { TaxConfig, TaxReport, PLReport } from '../services/financeApi';
import { taxConfigSchema, taxReportSchema, plReportSchema } from '../schemas/tax';
import { toLocalIso } from '../lib/date';

const reportClient = createOneKitClient(ReportServiceClient, 'finance');

export function useTaxConfig() {
  return useQuery({
    queryKey: ['tax-config'],
    queryFn: async () => {
      const res = await reportClient.listTaxConfig({});
      return z.array(taxConfigSchema).parse(res.configs) as TaxConfig[];
    },
  });
}

export function useTaxReport() {
  return useQuery({
    queryKey: ['tax-report'],
    queryFn: async () => {
      const res = await reportClient.getTaxReport({ period: 'month', date: '' });
      return taxReportSchema.parse(res) as TaxReport;
    },
  });
}

export function useCreateTaxConfig() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (data: { name: string; rate: number }) =>
      reportClient.createTaxConfig({
        name: data.name,
        rate: data.rate,
        type: 'sales',
        isDefault: false,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tax-config'] });
      qc.invalidateQueries({ queryKey: ['tax-report'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

/** Maps the short period token used by ProfitLoss.tsx to a report anchor date. */
function periodToAnchor(period: string): { period: string; date: string } {
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();

  if (period === 'last_month') {
    return { period: 'month', date: toLocalIso(new Date(y, m - 1, 15)) };
  }
  if (period === 'quarter' || period === 'year') {
    return { period, date: toLocalIso(now) };
  }
  // default: 'month'
  return { period: 'month', date: toLocalIso(now) };
}

/** Use this hook in ProfitLoss.tsx. Accepts a short period token (month/last_month/quarter/year). */
export function usePLByPeriod(period: string) {
  const anchor = periodToAnchor(period);
  return useQuery({
    queryKey: ['pl', 'filter', { period }],
    queryFn: async () => {
      const res = await reportClient.getPLReport(anchor);
      return plReportSchema.parse(res) as PLReport;
    },
    enabled: !!period,
  });
}
