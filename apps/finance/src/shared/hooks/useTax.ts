import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, TaxConfig } from '../services/financeApi';

export function useTaxConfig() {
  return useQuery({
    queryKey: ['tax-config'],
    queryFn: () => unwrap(financeApi.getTaxConfig()),
  });
}

export function useTaxReport() {
  return useQuery({
    queryKey: ['tax-report'],
    queryFn: () => unwrap(financeApi.getTaxReport()),
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
    queryFn: () => unwrap(financeApi.getPL(period)),
    enabled: !!period,
  });
}
