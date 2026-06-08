import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListRecurringInvoicesResponse } from '@voxire/contracts/finance/recurring_invoice/v1/list_recurring_invoices';
import type {
  CreateRecurringInvoiceRequest,
  CreateRecurringInvoiceResponse,
} from '@voxire/contracts/finance/recurring_invoice/v1/create_recurring_invoice';
import { safe } from '../../lib/api';

interface ListRecurringInvoicesParams {
  limit: number;
  offset: number;
  activeOnly?: boolean;
}

export const useListRecurringInvoices = (params: ListRecurringInvoicesParams) => {
  return useQuery<ListRecurringInvoicesResponse>({
    queryKey: ['recurring-invoices', 'filter', params],
    queryFn: async (): Promise<ListRecurringInvoicesResponse> => {
      const res = await unwrap(
        safe.get<ListRecurringInvoicesResponse>('/finance/recurring-invoices', { params })
      );
      return res as ListRecurringInvoicesResponse;
    },
  });
};

export const useCreateRecurringInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: CreateRecurringInvoiceRequest
    ): Promise<CreateRecurringInvoiceResponse> => {
      const source = data as CreateRecurringInvoiceRequest & {
        templateInvoiceId?: number;
        customerId?: number;
        startDate?: string;
        nextInvoiceDate?: string;
        endDate?: string;
        isActive?: boolean;
      };
      const res = await unwrap(
        safe.post<CreateRecurringInvoiceResponse>('/finance/recurring-invoices', {
          templateInvoiceId: source.templateInvoiceId ?? source.customerId,
          frequency: source.frequency,
          nextInvoiceDate: source.nextInvoiceDate ?? source.startDate,
          endDate: source.endDate,
          isActive: source.isActive ?? true,
        })
      );
      return res as CreateRecurringInvoiceResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recurring-invoices'] });
      toast.success(t('recurringInvoices.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

interface UpdateRecurringInvoiceVariables {
  id: number;
  data: Partial<CreateRecurringInvoiceRequest>;
}

export const useUpdateRecurringInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      variables: UpdateRecurringInvoiceVariables
    ): Promise<CreateRecurringInvoiceResponse> => {
      const res = await unwrap(
        safe.put<CreateRecurringInvoiceResponse>('/finance/recurring-invoices', {
          id: variables.id,
          ...variables.data,
        })
      );
      return res as CreateRecurringInvoiceResponse;
    },
    onSuccess: (_data, variables: UpdateRecurringInvoiceVariables) => {
      queryClient.invalidateQueries({ queryKey: ['recurring-invoices'] });
      queryClient.invalidateQueries({ queryKey: ['recurring-invoice', variables.id] });
      toast.success(t('recurringInvoices.updated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
