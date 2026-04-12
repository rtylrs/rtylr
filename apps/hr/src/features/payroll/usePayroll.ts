import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { hrApi } from '../../shared/services/hrApi';
import { payrollRowSchema } from '../../shared/schemas/payroll';

export function usePayroll(params?: { period?: string }) {
  return useQuery({
    queryKey: ['payroll', params],
    queryFn: async () => {
      const res = await unwrap(hrApi.listPayroll(params));
      const raw = res ?? [];
      return z.array(payrollRowSchema).parse(Array.isArray(raw) ? raw : []);
    },
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
