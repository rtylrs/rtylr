import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, Invoice } from '../services/financeApi';

export function useInvoices() {
  return useQuery({
    queryKey: ['invoices'],
    queryFn: () => financeApi.getInvoices(),
  });
}

export function useInvoiceDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['invoice', id],
    queryFn: () => financeApi.getInvoiceDetail(id!),
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Invoice>) => financeApi.createInvoice(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Invoice> }) => financeApi.updateInvoice(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => financeApi.deleteInvoice(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useUpdateInvoiceStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }) =>
      financeApi.updateInvoiceStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useRecordPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) => financeApi.recordPayment(id, amount),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}
