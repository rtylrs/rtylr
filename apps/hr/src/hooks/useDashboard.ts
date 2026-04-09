import { useQuery } from '@tanstack/react-query';
import { hrApi } from '../services/hrApi';

export function useHRDashboard() {
  return useQuery({
    queryKey: ['hr-dashboard'],
    queryFn: () => hrApi.getDashboard().then((r) => r.data),
  });
}
