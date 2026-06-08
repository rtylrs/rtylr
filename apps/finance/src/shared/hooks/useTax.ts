import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../lib/api';
import type { TaxConfig, TaxReport, PLReport } from '../services/financeApi';
import { taxConfigSchema, taxReportSchema, plReportSchema } from '../schemas/tax';

export function useTaxConfig() {
  return useQuery({
    queryKey: ['tax-config'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/tax/config'));
      const raw = res ?? [];
      return z.array(taxConfigSchema).parse(Array.isArray(raw) ? raw : []) as TaxConfig[];
    },
  });
}

export function useTaxReport() {
  return useQuery({
    queryKey: ['tax-report'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/tax/report'));
      return taxReportSchema.parse(res ?? {}) as TaxReport;
    },
  });
}

export function useCreateTaxConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TaxConfig): Promise<TaxConfig> =>
      unwrap(safe.post<TaxConfig>('/finance/tax/config', data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tax-config'] });
      qc.invalidateQueries({ queryKey: ['tax-report'] });
    },
    onError: () => toast.error('Failed to create tax config'),
  });
}

/** Maps the short period token used by ProfitLoss.tsx to a date range. */
function periodToRange(period: string): { from: string; to: string } {
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();

  if (period === 'last_month') {
    const firstOfLastMonth = new Date(y, m - 1, 1);
    const lastOfLastMonth = new Date(y, m, 0);
    return {
      from: firstOfLastMonth.toISOString().slice(0, 10),
      to: lastOfLastMonth.toISOString().slice(0, 10),
    };
  }
  if (period === 'quarter') {
    const quarterStart = new Date(y, Math.floor(m / 3) * 3, 1);
    return {
      from: quarterStart.toISOString().slice(0, 10),
      to: now.toISOString().slice(0, 10),
    };
  }
  if (period === 'year') {
    return {
      from: `${y}-01-01`,
      to: now.toISOString().slice(0, 10),
    };
  }
  // default: 'month'
  const firstOfMonth = new Date(y, m, 1);
  return {
    from: firstOfMonth.toISOString().slice(0, 10),
    to: now.toISOString().slice(0, 10),
  };
}

/** Use this hook in ProfitLoss.tsx. Accepts a short period token (month/last_month/quarter/year). */
export function usePLByPeriod(period: string) {
  const { from, to } = periodToRange(period);
  return useQuery({
    queryKey: ['pl', period],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/pl', { params: { from, to } }));
      return plReportSchema.parse(res ?? {}) as PLReport;
    },
    enabled: !!period,
  });
}
