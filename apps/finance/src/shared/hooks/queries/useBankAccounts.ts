import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListBankAccountsResponse } from '@voxire/contracts/finance/bank_account/v1/list_bank_accounts';
import type {
  CreateBankAccountRequest,
  CreateBankAccountResponse,
} from '@voxire/contracts/finance/bank_account/v1/create_bank_account';
import { safe } from '../../lib/api';

interface ListBankAccountsParams {
  limit: number;
  offset: number;
}

export const useListBankAccounts = (params: ListBankAccountsParams) => {
  return useQuery<ListBankAccountsResponse>({
    queryKey: ['bank-accounts', 'filter', params],
    queryFn: async (): Promise<ListBankAccountsResponse> => {
      const res = await unwrap(
        safe.get<ListBankAccountsResponse>('/finance/bank-accounts', { params })
      );
      return res as ListBankAccountsResponse;
    },
  });
};

export const useCreateBankAccount = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateBankAccountRequest): Promise<CreateBankAccountResponse> => {
      const source = data as CreateBankAccountRequest & {
        institution?: string;
        bankName?: string;
        accountType?: string;
        balance?: number;
        currency?: string;
        isConnected?: boolean;
      };
      const res = await unwrap(
        safe.post<CreateBankAccountResponse>('/finance/bank-accounts', {
          name: source.name,
          institution: source.institution ?? source.bankName,
          accountType: source.accountType,
          balance: source.balance ?? 0,
          currency: source.currency ?? 'USD',
          isConnected: source.isConnected ?? false,
        })
      );
      return res as CreateBankAccountResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bank-accounts'] });
      toast.success(t('bankAccounts.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
