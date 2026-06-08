import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { FLAG_KEYS, useFeatureFlag } from '@voxire/config';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet - define locally. Price-list endpoints are still
// being landed on the backend; hooks are flag-gated via APP_ERP so
// they are inert when the ERP surface is dark.
export interface PriceList {
  id: number;
  name: string;
  description: string;
  tier: 'Retail' | 'Wholesale' | 'VIP' | 'Custom';
  multiplier: number;
  activeCount: number;
}

export interface CreatePriceListInput {
  name: string;
  description?: string;
  tier: PriceList['tier'];
  multiplier: number;
}

interface ListPriceListsResponse {
  priceLists: PriceList[];
}

export const useListPriceLists = (): UseQueryResult<PriceList[]> => {
  const enabled = useFeatureFlag(FLAG_KEYS.APP_ERP);
  return useQuery<PriceList[]>({
    queryKey: ['erp-price-lists'],
    enabled,
    queryFn: async (): Promise<PriceList[]> => {
      const res = await unwrap(safe.get<ListPriceListsResponse>('/erp/price-lists'));
      return (res as ListPriceListsResponse).priceLists ?? [];
    },
  });
};

export const useCreatePriceList = (): UseMutationResult<PriceList, Error, CreatePriceListInput> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreatePriceListInput): Promise<PriceList> => {
      const res = await unwrap(
        safe.post<PriceList>('/erp/price-lists', {
          name: data.name,
          currency: 'USD',
          isDefault: false,
        })
      );
      return res as PriceList;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-price-lists'] });
      toast.success(t('erp.priceListCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
