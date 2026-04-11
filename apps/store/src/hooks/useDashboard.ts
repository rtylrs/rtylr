import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi } from '../services/storeService';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['store-dashboard'],
    queryFn: () => unwrap(storeApi.getDashboardStats()),
  });
}
