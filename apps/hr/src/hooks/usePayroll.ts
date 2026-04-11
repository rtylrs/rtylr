import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi } from '../services/hrApi';

export function usePayroll(params?: { period?: string }) {
  return useQuery({
    queryKey: ['payroll', params],
    queryFn: () => unwrap(hrApi.listPayroll(params)),
  });
}

export function useRunPayroll() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (period: string) => unwrap(hrApi.runPayroll(period)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payroll'] }),
  });
}

export function useMarkPayrollPaid() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(hrApi.markPayrollPaid(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['payroll'] }),
  });
}
