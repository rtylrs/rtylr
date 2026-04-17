import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useBusinesses() {
  return useQuery({
    queryKey: ['businesses'],
    queryFn: () => unwrap(flowApi.listBusinesses()).then((r) => r.businesses ?? []),
  });
}

export function useCreateBusiness() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color?: string; icon?: string }) =>
      unwrap(flowApi.createBusiness(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['businesses'] }),
  });
}

export function useSetWorkspaceBusiness() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ workspaceId, businessId }: { workspaceId: string; businessId: string | null }) =>
      unwrap(flowApi.setWorkspaceBusiness(workspaceId, businessId)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}
