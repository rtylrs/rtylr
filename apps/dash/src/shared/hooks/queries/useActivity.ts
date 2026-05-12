import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { activityEntrySchema } from '../../schemas/activity';

export type { ActivityEntry } from '../../schemas/activity';

export const useActivity = (limit = 100) => {
  return useQuery({
    queryKey: ['activity', limit],
    queryFn: async () => {
      const res = await unwrap(safe.get(`/activity?limit=${limit}`));
      const raw = res?.activities ?? res?.entries ?? res;
      return z.array(activityEntrySchema).parse(Array.isArray(raw) ? raw : []);
    },
    staleTime: 20_000,
  });
};

export const useAuditLog = (limit = 200) => {
  return useQuery({
    queryKey: ['audit', limit],
    queryFn: async () => {
      const res = await unwrap(safe.get(`/activity?limit=${limit}`));
      const raw = res?.activities ?? res?.entries ?? res;
      return z.array(activityEntrySchema).parse(Array.isArray(raw) ? raw : []);
    },
    staleTime: 20_000,
  });
};
