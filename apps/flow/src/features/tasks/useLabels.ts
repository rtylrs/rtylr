import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiLabel } from '../../shared/services/flowApi';

export function useLabels() {
  return useQuery<ApiLabel[]>({
    queryKey: ['labels'],
    queryFn: () => unwrap(flowApi.getLabels()),
  });
}

export function useCreateLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => unwrap(flowApi.createLabel(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
  });
}
