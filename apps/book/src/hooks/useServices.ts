import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export interface ServiceFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useServices(filters?: ServiceFilters) {
  return useQuery({
    queryKey: ['services', filters],
    queryFn: async () => {
      const res = await bookService.getServices(filters);
      return res.data?.data ?? res.data ?? [];
    },
  });
}

export function useService(id: number | undefined) {
  return useQuery({
    queryKey: ['service', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const res = await bookService.getService(id);
      return res.data?.data ?? res.data ?? null;
    },
  });
}

export function useCreateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => bookService.createService(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useUpdateService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) => bookService.updateService(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}

export function useDeleteService() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => bookService.deleteService(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['services'] }),
  });
}
