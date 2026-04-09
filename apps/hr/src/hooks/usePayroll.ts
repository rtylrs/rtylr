import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi } from '../services/hrApi';

export function usePayroll(params?: { period?: string }) {
  return useQuery({
    queryKey: ['payroll', params],
    queryFn: () => hrApi.listPayroll(params).then((r) => r.data || []),
  });
}

export function useRunPayroll() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (period: string) => hrApi.runPayroll(period),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payroll'] }),
  });
}

export function useMarkPayrollPaid() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => hrApi.markPayrollPaid(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payroll'] }),
  });
}
