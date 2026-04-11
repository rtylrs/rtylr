import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ReorderSuggestion } from '../../shared/services/erpApi';

export function useReorderSuggestions() {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['reorder-suggestions'],
    queryFn: () => unwrap(erpApi.getReorderSuggestions()),
  });
}
