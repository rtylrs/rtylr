import { useQuery } from '@tanstack/react-query';
import { flowPresenceClient } from '../../shared/lib/clients';
import { presenceResponseSchema } from '../../shared/schemas/presence';

/**
 * Business-wide "who's in today," gated server-side by view_directory
 * (onion PR #267) - a caller without it gets a real 403 here, the same
 * disabled-with-a-reason rule Directory and Team/Reports already use.
 */
export function usePresence() {
  return useQuery({
    queryKey: ['presence-today'],
    queryFn: async () => {
      const res = await flowPresenceClient.getPresence({});
      return presenceResponseSchema.parse(res);
    },
  });
}
