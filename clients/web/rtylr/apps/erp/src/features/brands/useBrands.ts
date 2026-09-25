import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { BrandServiceClient } from '@voxire/voxie/erp/brand/v1/client';
import type {
  Brand,
  CreateBrandRequest,
  CreateBrandResponse,
  DeleteBrandResponse,
  UpdateBrandRequest,
  UpdateBrandResponse,
} from '@voxire/voxie/erp/brand/v1/types';
import { ERP_LIST_LIMIT } from '../../shared/lib/erpListBounds';

const brandClient = createOneKitClient(BrandServiceClient, 'erp');

export type { Brand };

export async function fetchAllBrands(): Promise<Brand[]> {
  const brands: Brand[] = [];
  let offset = 0;

  while (true) {
    const response = await brandClient.listBrands({ limit: ERP_LIST_LIMIT, offset });
    const page = response.brands ?? [];
    brands.push(...page);

    const hasMore =
      response.hasMore ??
      (response.total !== undefined
        ? brands.length < response.total
        : page.length === ERP_LIST_LIMIT);
    if (!hasMore || page.length === 0) return brands;
    offset += page.length;
  }
}

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: fetchAllBrands,
  });
}

export function useCreateBrand() {
  const qc = useQueryClient();
  return useMutation<CreateBrandResponse, Error, CreateBrandRequest>({
    mutationFn: (data) => brandClient.createBrand(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
    onError: () => toast.error('Failed to create brand'),
  });
}

export function useDeleteBrand() {
  const qc = useQueryClient();
  return useMutation<DeleteBrandResponse, Error, string>({
    mutationFn: (id) => brandClient.deleteBrand({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
    onError: () => toast.error('Failed to delete brand'),
  });
}

export function useUpdateBrand() {
  const qc = useQueryClient();
  return useMutation<UpdateBrandResponse, Error, UpdateBrandRequest>({
    mutationFn: (data) => brandClient.updateBrand(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
    onError: () => toast.error('Failed to update brand'),
  });
}
