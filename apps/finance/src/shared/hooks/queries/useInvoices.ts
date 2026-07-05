import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  InvoicesServiceClient,
  type ListInvoicesResponse,
} from '@voxire/voxie/finance/invoices/v1/service_client';

const invoicesClient = createOneKitClient(InvoicesServiceClient, 'finance');

interface ListInvoicesParams {
  limit: number;
  offset: number;
  status?: string;
  search?: string;
  dateFrom?: string;
  dateTo?: string;
}

export const useListInvoices = (params: ListInvoicesParams) => {
  return useQuery<ListInvoicesResponse>({
    queryKey: ['invoices', 'filter', params],
    queryFn: (): Promise<ListInvoicesResponse> =>
      invoicesClient.listInvoices({
        status: params.status ?? '',
        search: params.search ?? '',
        dateFrom: params.dateFrom ?? '',
        dateTo: params.dateTo ?? '',
        page: Math.floor(params.offset / Math.max(1, params.limit)) + 1,
        pageSize: params.limit,
      }),
  });
};

export interface SendInvoiceRequest {
  invoiceId: string;
  // finance.invoices.v1.SendInvoice takes only the invoice id - it doesn't
  // support custom recipient/subject/body (the legacy REST endpoint it
  // replaces silently ignored these same fields too).
  to: string;
  cc?: string;
  subject: string;
  body: string;
}

export interface SendInvoiceResponse {
  message: string;
}

export const useSendInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: SendInvoiceRequest): Promise<SendInvoiceResponse> =>
      invoicesClient.sendInvoice({ id: data.invoiceId }),
    onSuccess: (_data: SendInvoiceResponse, variables: SendInvoiceRequest) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      toast.success(t('invoices.sent', { to: variables.to }));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};
