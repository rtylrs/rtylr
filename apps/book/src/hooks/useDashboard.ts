import { useQuery } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const res = await bookService.getDashboardStats();
      return res.data?.data ?? res.data ?? null;
    },
  });
}

export function useTodayAppointments() {
  return useQuery({
    queryKey: ['dashboard', 'today'],
    queryFn: async () => {
      const res = await bookService.getTodayAppointments();
      return res.data?.data ?? res.data ?? [];
    },
  });
}
