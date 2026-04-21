import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type {
  ListInvoicesResponse,
  GetInvoiceResponse,
  CreateInvoiceRequest,
  CreateInvoiceResponse,
  UpdateInvoiceRequest,
  UpdateInvoiceResponse,
  DeleteInvoiceResponse,
  RecordPaymentRequest,
  RecordPaymentResponse,
  UpdateInvoiceStatusResponse,
} from '@voxire/contracts/finance/v1/invoices';
import { safe } from '../../lib/api';

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
    queryFn: async (): Promise<ListInvoicesResponse> => {
      const res = await unwrap(safe.get<ListInvoicesResponse>('/finance/invoices', { params }));
      return res as ListInvoicesResponse;
    },
  });
};

export const useGetInvoice = (id: number | undefined) => {
  return useQuery<GetInvoiceResponse>({
    queryKey: ['invoice', id],
    queryFn: async (): Promise<GetInvoiceResponse> => {
      const res = await unwrap(safe.get<GetInvoiceResponse>(`/finance/invoices/${id}`));
      return res as GetInvoiceResponse;
    },
    enabled: !!id,
  });
};

export const useCreateInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateInvoiceRequest): Promise<CreateInvoiceResponse> => {
      const res = await unwrap(safe.post<CreateInvoiceResponse>('/finance/invoices', data));
      return res as CreateInvoiceResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      toast.success(t('invoices.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateInvoiceRequest): Promise<UpdateInvoiceResponse> => {
      const res = await unwrap(
        safe.put<UpdateInvoiceResponse>(`/finance/invoices/${data.id}`, data)
      );
      return res as UpdateInvoiceResponse;
    },
    onSuccess: (_data: UpdateInvoiceResponse, variables: UpdateInvoiceRequest) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['invoice', variables.id] });
      toast.success(t('invoices.updated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<DeleteInvoiceResponse> => {
      const res = await unwrap(safe.delete<DeleteInvoiceResponse>(`/finance/invoices/${id}`));
      return res as DeleteInvoiceResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      toast.success(t('invoices.deleted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useRecordInvoicePayment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: RecordPaymentRequest): Promise<RecordPaymentResponse> => {
      const res = await unwrap(
        safe.post<RecordPaymentResponse>(`/finance/invoices/${data.invoiceId}/payment`, data)
      );
      return res as RecordPaymentResponse;
    },
    onSuccess: (_data: RecordPaymentResponse, variables: RecordPaymentRequest) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['invoice', variables.invoiceId] });
      toast.success(t('invoices.paymentRecorded'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

interface SetInvoiceStatusVariables {
  id: number;
  status: string;
}

export const useSetInvoiceStatus = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: SetInvoiceStatusVariables): Promise<UpdateInvoiceStatusResponse> => {
      const res = await unwrap(
        safe.put<UpdateInvoiceStatusResponse>(`/finance/invoices/${data.id}/status`, {
          status: data.status,
        })
      );
      return res as UpdateInvoiceStatusResponse;
    },
    onSuccess: (_data: UpdateInvoiceStatusResponse, variables: SetInvoiceStatusVariables) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['invoice', variables.id] });
      toast.success(t('invoices.statusUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
