import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { RecurringInvoiceServiceClient } from '@voxire/voxie/finance/recurring_invoice/v1/client';
import { recurringInvoiceSchema } from '../../shared/schemas/recurringInvoices';

const recurringClient = createOneKitClient(RecurringInvoiceServiceClient, 'finance');

interface UseRecurringInvoicesParams {
  page?: number;
  pageSize?: number;
}

export interface RecurringInvoiceListResult {
  invoices: z.infer<typeof recurringInvoiceSchema>[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function useRecurringInvoices({ page = 1, pageSize = 50 }: UseRecurringInvoicesParams = {}) {
  return useQuery<RecurringInvoiceListResult>({
    queryKey: ['recurring-invoices', { page, pageSize }],
    queryFn: async (): Promise<RecurringInvoiceListResult> => {
      const res = await recurringClient.listRecurringInvoices({ page, pageSize });
      return {
        invoices: z.array(recurringInvoiceSchema).parse(
          (res.recurringInvoices ?? []).map((invoice) => ({
            ...invoice,
            // The Go JSON encoder omits false for an omitempty bool.
            isActive: invoice.isActive ?? false,
          }))
        ),
        hasMore: res.hasMore ?? false,
        page: res.page ?? page,
        pageSize: res.pageSize ?? pageSize,
      };
    },
  });
}

interface CreateRecurringPayload {
  templateInvoiceId: string;
  frequency: string;
  nextInvoiceDate: string;
  endDate?: string | null;
  isActive?: boolean;
}

export function useCreateRecurringInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (data: CreateRecurringPayload) =>
      recurringClient.createRecurringInvoice({
        templateInvoiceId: data.templateInvoiceId,
        frequency: data.frequency,
        nextInvoiceDate: data.nextInvoiceDate,
        endDate: data.endDate ?? undefined,
        isActive: data.isActive ?? true,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['recurring-invoices'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useSetRecurringStatus() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'active' | 'paused' }) =>
      status === 'paused'
        ? recurringClient.pauseRecurringInvoice({ id })
        : recurringClient.resumeRecurringInvoice({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring-invoices'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}
