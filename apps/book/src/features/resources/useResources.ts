import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type {
  CreateResourceRequest,
  UpdateResourceRequest,
} from '@voxire/contracts/book/v1/resources';
import { bookService } from '../../shared/services/bookService';

export interface ResourceFilters {
  search?: string;
  type?: string;
  limit?: number;
  offset?: number;
}

export function useResources(filters?: ResourceFilters) {
  return useQuery({
    queryKey: ['resources', filters],
    queryFn: () => unwrap(bookService.getResources(filters)),
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
