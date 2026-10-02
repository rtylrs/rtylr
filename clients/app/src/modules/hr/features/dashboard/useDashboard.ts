import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi } from '../../shared/services/hrApi';
import { hrDashboardStatsSchema, getDashboardResponseSchema } from '../../shared/schemas/dashboard';
import type { DashboardWidget } from '../../shared/schemas/dashboard';

export function useHRDashboard() {
  return useQuery({
    queryKey: ['hr', 'hr-dashboard'],
    queryFn: async () => {
      const res = await unwrap(hrApi.getDashboard());
      return hrDashboardStatsSchema.parse(res);
    },
  });
}

export function useGetDashboard() {
  return useQuery({
    queryKey: ['hr', 'hr-dashboard-widgets'],
    queryFn: async () => {
      const res = await unwrap(hrApi.getDashboardWidgets());
      return getDashboardResponseSchema.parse(res);
    },
    staleTime: 1000 * 60,
  });
}

export function useUpdateWidgetLayout(): ReturnType<
  typeof useMutation<{ success: boolean }, Error, DashboardWidget[]>
> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (widgets: DashboardWidget[]) => unwrap(hrApi.updateWidgetLayout(widgets)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hr', 'hr-dashboard'] });
      qc.invalidateQueries({ queryKey: ['hr', 'hr-dashboard-widgets'] });
    },
  });
}
