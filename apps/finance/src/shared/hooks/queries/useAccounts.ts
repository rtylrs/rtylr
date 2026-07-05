import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListAccountsResponse } from '@voxire/voxie/finance/account/v1/list_accounts';
import type {
  CreateAccountRequest,
  CreateAccountResponse,
} from '@voxire/voxie/finance/account/v1/create_account';
import { safe } from '../../lib/api';

interface ListAccountsParams {
  limit: number;
  offset: number;
  type?: string;
}

export const useListAccounts = (params: ListAccountsParams) => {
  return useQuery<ListAccountsResponse>({
    queryKey: ['accounts', 'filter', params],
    queryFn: async (): Promise<ListAccountsResponse> => {
      const res = await unwrap(safe.get<ListAccountsResponse>('/finance/accounts', { params }));
      return res as ListAccountsResponse;
    },
  });
};

export const useCreateAccount = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateAccountRequest): Promise<CreateAccountResponse> => {
      const res = await unwrap(safe.post<CreateAccountResponse>('/finance/accounts', data));
      return res as CreateAccountResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['accounts'] });
      toast.success(t('accounts.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
