import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Brand } from '../../shared/services/erpApi';
import { brandSchema } from '../../shared/schemas/brands';

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listBrands());
      return z.array(brandSchema).parse(Array.isArray(res) ? res : []) as Brand[];
    },
  });
}
