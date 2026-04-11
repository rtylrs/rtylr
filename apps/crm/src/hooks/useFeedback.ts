import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { crmApi } from '../services/crmApi';

export function useFeedback() {
  return useQuery({
    queryKey: ['feedback'],
    queryFn: () => unwrap(crmApi.getFeedback()),
  });
}
