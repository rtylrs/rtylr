import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { rest } from '../../shared/lib/clients';
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
      // /flow/presence/today has no .onk definition, so there is no generated
      // client for it. Parsed at the boundary below (#894).
      const res = await unwrap(rest.get<unknown>('/flow/presence/today'));
      return presenceResponseSchema.parse(res);
    },
  });
}
