import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import type {
  CreateResourceRequest,
  UpdateResourceRequest,
} from '@voxire/contracts/book/v1/resources';
import { bookService } from '../../shared/services/bookService';
import { resourceSchema } from '../../shared/schemas/resources';

export interface ResourceFilters {
  search?: string;
  type?: string;
  limit?: number;
  offset?: number;
}

export function useResources(filters?: ResourceFilters) {
  return useQuery({
    queryKey: ['resources', filters],
    queryFn: async () => {
      const res = await unwrap(bookService.getResources(filters));
      const raw = (res as any)?.resources ?? res;
      return z.array(resourceSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateResource() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateResourceRequest) => unwrap(bookService.createResource(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['resources'] }),
  });
}

export function useUpdateResource() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateResourceRequest }) =>
      unwrap(bookService.updateResource(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['resources'] }),
  });
}

export function useDeleteResource() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(bookService.deleteResource(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['resources'] }),
  });
}
