import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ActivityItem } from '../services/erpApi';

export function useActivity(limit = 20) {
  return useQuery<ActivityItem[]>({
    queryKey: ['activity', limit],
    queryFn: () => unwrap(erpApi.listActivity(limit)),
  });
}
