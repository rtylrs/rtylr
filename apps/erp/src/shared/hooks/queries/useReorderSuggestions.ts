import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';

// No proto type yet — define locally.
export interface ReorderSuggestion {
  productId: number;
  name: string;
  sku: string;
  currentStock: number;
  avgDailySales: number;
  suggestedQuantity: number;
  leadTimeDays: number;
  confidence: 'high' | 'medium' | 'low';
}

interface ReorderSuggestionsResponse {
  suggestions: ReorderSuggestion[];
}

export const useReorderSuggestions = (): UseQueryResult<ReorderSuggestion[]> => {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['erp-reorder-suggestions'],
    queryFn: async (): Promise<ReorderSuggestion[]> => {
      const res = await unwrap(
        safe.get<ReorderSuggestionsResponse | ReorderSuggestion[]>('/erp/reorder-suggestions')
      );
      if (Array.isArray(res)) return res;
      return (res as ReorderSuggestionsResponse).suggestions ?? [];
    },
  });
};
