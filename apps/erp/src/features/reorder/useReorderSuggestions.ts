import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ReorderSuggestion } from '../../shared/services/erpApi';
import { reorderSuggestionSchema } from '../../shared/schemas/reorder';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useReorderSuggestions() {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['reorder-suggestions'],
    queryFn: async () => {
      const res = await unwrap(erpApi.getReorderSuggestions());
      return z
        .array(reorderSuggestionSchema)
        .parse(unwrapList(res, 'suggestions', 'items')) as ReorderSuggestion[];
    },
  });
}
