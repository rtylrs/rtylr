import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { crmApi } from '../../shared/services/crmApi';

export function useCRMDashboard() {
  return useQuery({
    queryKey: ['crm-dashboard'],
    queryFn: () => unwrap(crmApi.getDashboard()),
  });
}
