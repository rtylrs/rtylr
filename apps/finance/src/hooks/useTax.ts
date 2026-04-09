import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, TaxConfig } from '../services/financeApi';

export function useTaxConfig() {
  return useQuery({
    queryKey: ['tax-config'],
    queryFn: () => financeApi.getTaxConfig(),
  });
}

export function useTaxReport() {
  return useQuery({
    queryKey: ['tax-report'],
    queryFn: () => financeApi.getTaxReport(),
  });
}

export function useCreateTaxConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: TaxConfig) => financeApi.createTaxConfig(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tax-config'] });
      qc.invalidateQueries({ queryKey: ['tax-report'] });
    },
  });
}

export function usePL(period: string) {
  return useQuery({
    queryKey: ['pl', period],
    queryFn: () => financeApi.getPL(period),
    enabled: !!period,
  });
}
