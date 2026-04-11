import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ExpiryItem } from '../../shared/services/erpApi';

export function useExpiry() {
  return useQuery<ExpiryItem[]>({
    queryKey: ['expiry'],
    queryFn: () => unwrap(erpApi.getExpiryItems()),
  });
}
