import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from '@tanstack/react-query';
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

/**
 * A mutation whose transport-level `Idempotency-Key` rides in its variables.
 *
 * The key identifies the user action, not the request. Minting it inside
 * `mutationFn` would give every retry of one submit a different key, so both
 * would execute and the server's request fingerprint could not tell a retry
 * from two real actions.
 */
export type Keyed<TVariables> = TVariables & { idempotencyKey: string };

/**
 * Wraps a mutation so every `mutate`/`mutateAsync` call carries its own
 * `Idempotency-Key`, minted once when the operator acts and reused by every
 * attempt of that action.
 *
 * The key travels in the variables because that is the one thing react-query
 * hands back unchanged to `mutationFn` on each retry. A later, separate submit
 * mints a different key, so the server can still tell the two apart.
 */
export function withIdempotencyKey<TData, TError, TVariables>(
  mutation: UseMutationResult<TData, TError, Keyed<TVariables>>
) {
  return {
    ...mutation,
    mutate: (variables: TVariables, options?: Parameters<typeof mutation.mutate>[1]) =>
      mutation.mutate({ ...variables, idempotencyKey: crypto.randomUUID() }, options),
    mutateAsync: (variables: TVariables, options?: Parameters<typeof mutation.mutateAsync>[1]) =>
      mutation.mutateAsync({ ...variables, idempotencyKey: crypto.randomUUID() }, options),
  };
}

export const useSendInvoice = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  const mutation = useMutation<SendInvoiceResponse, Error, Keyed<SendInvoiceRequest>>({
    mutationFn: async (data: Keyed<SendInvoiceRequest>): Promise<SendInvoiceResponse> => {
      const { idempotencyKey, invoiceId } = data;
      // The generated client already accepts per-call RequestOptions, so the
      // key reaches the wire without a contract change. Without it the server
      // answers this POST with 428 and the invoice is never sent.
      const res = await invoicesClient.sendInvoice(
        { id: invoiceId },
        { headers: { 'Idempotency-Key': idempotencyKey } }
      );
      return { message: res.message ?? '' };
    },
    onSuccess: (_data: SendInvoiceResponse, variables: Keyed<SendInvoiceRequest>) => {
      queryClient.invalidateQueries({ queryKey: ['invoices'] });
      queryClient.invalidateQueries({ queryKey: ['invoice', variables.invoiceId] });
      toast.success(t('invoices.sent'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });

  return withIdempotencyKey(mutation);
};
