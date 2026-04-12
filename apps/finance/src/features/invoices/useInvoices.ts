import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';

export function useInvoices() {
  return useQuery({
    queryKey: ['invoices'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getInvoices());
      const raw = res ?? [];
      return z.array(invoiceSchema).parse(Array.isArray(raw) ? raw : []) as Invoice[];
    },
  });
}

export function useInvoiceDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['invoice', id],
    queryFn: async () => {
      const res = await unwrap(financeApi.getInvoiceDetail(id!));
      return invoiceSchema.parse(res) as Invoice;
    },
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Invoice>) => unwrap(financeApi.createInvoice(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Invoice> }) =>
      unwrap(financeApi.updateInvoice(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(financeApi.deleteInvoice(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useUpdateInvoiceStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }) =>
      unwrap(financeApi.updateInvoiceStatus(id, status)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}

export function useRecordPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      unwrap(financeApi.recordPayment(id, amount)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
  });
}
