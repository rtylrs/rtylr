import { useQuery } from '@tanstack/react-query';
import { crmApi } from '../services/crmApi';

export function useFeedback() {
  return useQuery({
    queryKey: ['feedback'],
    queryFn: () => crmApi.getFeedback(),
  });
}
