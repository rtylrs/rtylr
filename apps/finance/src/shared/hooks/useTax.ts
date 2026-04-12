import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, TaxConfig, TaxReport, PLReport } from '../services/financeApi';
import { taxConfigSchema, taxReportSchema, plReportSchema } from '../schemas/tax';

export function useTaxConfig() {
  return useQuery({
    queryKey: ['tax-config'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getTaxConfig());
      const raw = res ?? [];
      return z.array(taxConfigSchema).parse(Array.isArray(raw) ? raw : []) as TaxConfig[];
    },
  });
}

export function useTaxReport() {
  return useQuery({
    queryKey: ['tax-report'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getTaxReport());
      return taxReportSchema.parse(res ?? {}) as TaxReport;
    },
  });
}

export function useCreateTaxConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TaxConfig) => unwrap(financeApi.createTaxConfig(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tax-config'] });
      qc.invalidateQueries({ queryKey: ['tax-report'] });
    },
  });
}

export function usePL(period: string) {
  return useQuery({
    queryKey: ['pl', period],
    queryFn: async () => {
      const res = await unwrap(financeApi.getPL(period));
      return plReportSchema.parse(res ?? {}) as PLReport;
    },
    enabled: !!period,
  });
}
