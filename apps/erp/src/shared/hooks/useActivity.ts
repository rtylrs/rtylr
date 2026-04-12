import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ActivityItem } from '../services/erpApi';
import { activityItemSchema } from '../schemas/activity';

export function useActivity(limit = 20) {
  return useQuery<ActivityItem[]>({
    queryKey: ['activity', limit],
    queryFn: async () => {
      const res = await unwrap(erpApi.listActivity(limit));
      return z.array(activityItemSchema).parse(Array.isArray(res) ? res : []) as ActivityItem[];
    },
  });
}
