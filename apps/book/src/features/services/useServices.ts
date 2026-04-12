import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import type {
  CreateBookServiceRequest,
  UpdateBookServiceRequest,
} from '@voxire/contracts/book/v1/services';
import { bookService } from '../../shared/services/bookService';
import { bookServiceSchema } from '../../shared/schemas/services';

export interface ServiceFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useServices(filters?: ServiceFilters) {
  return useQuery({
    queryKey: ['services', filters],
    queryFn: async () => {
      const res = await unwrap(bookService.getServices(filters));
      const raw = (res as Record<string, unknown>)?.services ?? res;
      return z.array(bookServiceSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useService(id: number | undefined) {
  return useQuery({
    queryKey: ['service', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await unwrap(bookService.getService(id!));
      return bookServiceSchema.parse(res);
    },
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
