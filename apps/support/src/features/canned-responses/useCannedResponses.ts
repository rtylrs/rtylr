import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { supportService } from '../../shared/services/supportService';
import { cannedResponseSchema } from '../../shared/schemas/cannedResponses';

export function useCannedResponses(params?: { search?: string }) {
  return useQuery({
    queryKey: ['canned-responses', params],
    queryFn: async () => {
      const res = await unwrap(supportService.getCannedResponses(params));
      const raw = res ?? [];
      return z.array(cannedResponseSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCannedResponse(id: number) {
  return useQuery({
    queryKey: ['canned-response', id],
    queryFn: async () => {
      const res = await unwrap(supportService.getCannedResponse(id));
      return cannedResponseSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useCreateCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { title: string; body: string }) =>
      unwrap(supportService.createCannedResponse(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
    },
  });
}

export function useUpdateCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: { title?: string; body?: string } }) =>
      unwrap(supportService.updateCannedResponse(id, data)),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
      qc.invalidateQueries({ queryKey: ['canned-response', vars.id] });
    },
  });
}

export function useDeleteCannedResponse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(supportService.deleteCannedResponse(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['canned-responses'] });
    },
  });
}
