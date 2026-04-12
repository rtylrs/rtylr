import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { bookService } from '../../shared/services/bookService';
import { bookDashboardStatsSchema } from '../../shared/schemas/dashboard';
import { appointmentSchema } from '../../shared/schemas/appointments';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: async () => {
      const res = await unwrap(bookService.getDashboardStats());
      return bookDashboardStatsSchema.parse(res);
    },
  });
}

export function useTodayAppointments() {
  return useQuery({
    queryKey: ['dashboard', 'today'],
    queryFn: async () => {
      const res = await unwrap(bookService.getTodayAppointments());
      const raw = (res as any)?.appointments ?? res;
      return z.array(appointmentSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
