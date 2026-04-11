import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export interface WaitlistFilters {
  search?: string;
  status?: string;
  limit?: number;
  offset?: number;
}

export function useWaitlist(filters?: WaitlistFilters) {
  return useQuery({
    queryKey: ['waitlist', filters],
    queryFn: async () => {
      const res = await bookService.getWaitlist(filters);
      return res.data?.data ?? res.data ?? [];
    },
  });
}

export function useAddToWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => bookService.addToWaitlist(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}

export function useNotifyWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => bookService.notifyWaitlist(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}

export function useRemoveFromWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => bookService.removeFromWaitlist(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}
