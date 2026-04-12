import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import type {
  CreateBookServiceRequest,
  UpdateBookServiceRequest,
} from '@voxire/contracts/book/v1/services';
import { bookService } from '../../shared/services/bookService';

export interface ServiceFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useServices(filters?: ServiceFilters) {
  return useQuery({
    queryKey: ['services', filters],
    queryFn: () => unwrap(bookService.getServices(filters)),
  });
}

export function useService(id: number | undefined) {
  return useQuery({
    queryKey: ['service', id],
    enabled: !!id,
    queryFn: () => unwrap(bookService.getService(id!)),
  });
}

export function useCreateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateBookServiceRequest) => unwrap(bookService.createService(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useUpdateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateBookServiceRequest }) =>
      unwrap(bookService.updateService(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useDeleteService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(bookService.deleteService(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}
