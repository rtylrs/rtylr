import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Brand } from '../../shared/services/erpApi';

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: () => unwrap(erpApi.listBrands()),
  });
}
