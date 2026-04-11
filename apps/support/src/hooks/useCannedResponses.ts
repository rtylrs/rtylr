import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useCannedResponses(params?: { search?: string }) {
  return useQuery({
    queryKey: ['canned-responses', params],
    queryFn: () => supportService.getCannedResponses(params).then((r) => r.data),
  });
}

export function useCannedResponse(id: number) {
  return useQuery({
    queryKey: ['canned-response', id],
    queryFn: () => supportService.getCannedResponse(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useCreateCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { title: string; body: string }) =>
      supportService.createCannedResponse(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
    },
  });
}

export function useUpdateCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: { title?: string; body?: string } }) =>
      supportService.updateCannedResponse(id, data).then((r) => r.data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
      qc.invalidateQueries({ queryKey: ['canned-response', vars.id] });
    },
  });
}

export function useDeleteCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => supportService.deleteCannedResponse(id).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
    },
  });
}
