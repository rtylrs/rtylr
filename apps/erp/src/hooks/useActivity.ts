import { useQuery } from '@tanstack/react-query';
import { erpApi, ActivityItem } from '../services/erpApi';
import { mockActivity } from '../lib/mockData';

export function useActivity(limit = 20) {
  return useQuery<ActivityItem[]>({
    queryKey: ['activity', limit],
    queryFn: async () => {
      try {
        const res = await erpApi.listActivity(limit);
        if (res.data && res.data.length) return res.data;
        return mockActivity.slice(0, limit);
      } catch {
        return mockActivity.slice(0, limit);
      }
    },
  });
}
