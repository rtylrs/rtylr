import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import type { Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useInvoices() {
  return useQuery<Invoice[]>({
    queryKey: ['invoices'],
    queryFn: async (): Promise<Invoice[]> => {
      const res = await unwrap(safe.get<unknown>('/finance/invoices'));
      return z.array(invoiceSchema).parse(unwrapList(res, 'invoices')) as Invoice[];
    },
  });
}

export function useInvoiceDetail(id: string | undefined) {
  return useQuery<Invoice>({
    queryKey: ['invoice', id],
    queryFn: async (): Promise<Invoice> => {
      const res = await unwrap(safe.get<unknown>(`/finance/invoices/${id}`));
      return invoiceSchema.parse(res) as Invoice;
    },
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Invoice>): Promise<Invoice> =>
      unwrap(safe.post<Invoice>('/finance/invoices', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to create invoice'),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Invoice> }): Promise<Invoice> =>
      unwrap(safe.put<Invoice>(`/finance/invoices/${id}`, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to update invoice'),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string): Promise<void> => unwrap(safe.delete<void>(`/finance/invoices/${id}`)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to delete invoice'),
  });
}

export function useUpdateInvoiceStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }): Promise<void> =>
      unwrap(safe.put<void>(`/finance/invoices/${id}/status`, { status })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to update invoice status'),
  });
}

export function useRecordPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      amount,
      paymentDate,
      method,
    }: {
      id: string;
      amount: number;
      paymentDate?: string;
      method?: string;
    }): Promise<void> =>
      unwrap(
        safe.post<void>('/finance/invoices/payment', {
          invoiceId: Number(id),
          amount,
          paymentDate,
          method,
        })
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to record payment'),
  });
}
