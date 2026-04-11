import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { hrApi } from '../../shared/services/hrApi';

export function useHRDashboard() {
  return useQuery({
    queryKey: ['hr-dashboard'],
    queryFn: () => unwrap(hrApi.getDashboard()),
  });
}
