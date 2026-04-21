import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet — define locally.
export interface Brand {
  id: number;
  name: string;
  logoUrl?: string;
  description?: string;
  productCount: number;
}

export interface CreateBrandInput {
  name: string;
  logoUrl?: string;
  description?: string;
}

interface ListBrandsResponse {
  brands: Brand[];
}

export const useListBrands = (): UseQueryResult<Brand[]> => {
  return useQuery<Brand[]>({
    queryKey: ['erp-brands'],
    queryFn: async (): Promise<Brand[]> => {
      const res = await unwrap(safe.get<ListBrandsResponse>('/erp/brands'));
      return (res as ListBrandsResponse).brands ?? [];
    },
  });
};

export const useCreateBrand = (): UseMutationResult<Brand, Error, CreateBrandInput> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateBrandInput): Promise<Brand> => {
      const res = await unwrap(safe.post<Brand>('/erp/brands', data));
      return res as Brand;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-brands'] });
      toast.success(t('erp.brandCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
