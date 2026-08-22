import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import type { ListInvoicesResponse } from '@voxire/voxie/finance/invoices/v1/types';

const invoicesClient = createOneKitClient(InvoicesServiceClient, 'finance');

export interface ListInvoicesParams {
  page?: number;
  pageSize?: number;
  status?: string;
  search?: string;
  dateFrom?: string;
  dateTo?: string;
}

export const useListInvoices = (params: ListInvoicesParams) => {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;

  return useQuery<ListInvoicesResponse>({
    queryKey: [
      'invoices',
      'filter',
      {
        page,
        pageSize,
        status: params.status ?? '',
        search: params.search ?? '',
        dateFrom: params.dateFrom ?? '',
        dateTo: params.dateTo ?? '',
      },
    ],
    queryFn: (): Promise<ListInvoicesResponse> =>
      invoicesClient.listInvoices({
        status: params.status ?? '',
        search: params.search ?? '',
        dateFrom: params.dateFrom ?? '',
        dateTo: params.dateTo ?? '',
        page,
        pageSize,
      }),
  });
};

export interface SendInvoiceRequest {
  invoiceId: string;
}

export interface SendInvoiceResponse {
  message: string;
}

export const useSendInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: SendInvoiceRequest): Promise<SendInvoiceResponse> => {
      const res = await invoicesClient.sendInvoice({ id: data.invoiceId });
      return { message: res.message ?? '' };
    },
    onSuccess: (_data: SendInvoiceResponse, variables: SendInvoiceRequest) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['invoice', variables.invoiceId] });
      toast.success(t('invoices.sent'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};
