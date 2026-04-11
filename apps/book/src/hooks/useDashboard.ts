import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { bookService } from '../services/bookService';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['dashboard', 'stats'],
    queryFn: () => unwrap(bookService.getDashboardStats()),
  });
}

export function useTodayAppointments() {
  return useQuery({
    queryKey: ['dashboard', 'today'],
    queryFn: () => unwrap(bookService.getTodayAppointments()),
  });
}
