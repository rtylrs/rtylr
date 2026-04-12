import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import type { CreateWaitlistEntryRequest } from '@voxire/contracts/book/v1/waitlist';
import { bookService } from '../../shared/services/bookService';
import { waitlistEntrySchema } from '../../shared/schemas/waitlist';

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
      const res = await unwrap(bookService.getWaitlist(filters));
      const raw = (res as Record<string, unknown>)?.entries ?? res;
      return z.array(waitlistEntrySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useAddToWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWaitlistEntryRequest) => unwrap(bookService.addToWaitlist(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}

export function useNotifyWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(bookService.notifyWaitlist(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}

export function useRemoveFromWaitlist() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(bookService.removeFromWaitlist(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waitlist'] }),
  });
}
