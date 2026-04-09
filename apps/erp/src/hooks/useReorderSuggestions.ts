import { useQuery } from '@tanstack/react-query';
import { erpApi, ReorderSuggestion } from '../services/erpApi';
import { mockReorderSuggestions } from '../lib/mockData';

export function useReorderSuggestions() {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['reorder-suggestions'],
    queryFn: async () => {
      try {
        const res = await erpApi.getReorderSuggestions();
        if (res.data && res.data.length) return res.data;
        return mockReorderSuggestions;
      } catch {
        return mockReorderSuggestions;
      }
    },
  });
}
