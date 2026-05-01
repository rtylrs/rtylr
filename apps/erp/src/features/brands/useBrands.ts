import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, Brand } from '../../shared/services/erpApi';
import { brandSchema } from '../../shared/schemas/brands';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listBrands());
      return z.array(brandSchema).parse(unwrapList(res, 'brands')) as Brand[];
    },
  });
}

export function useCreateBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Brand>) => unwrap(erpApi.createBrand(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
  });
}

export function useDeleteBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteBrand(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
  });
}
