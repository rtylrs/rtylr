import { useQuery } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export interface CalendarFilters {
  start?: string;
  end?: string;
  staffId?: number;
}

export function useCalendarEvents(filters?: CalendarFilters) {
  return useQuery({
    queryKey: ['calendar', filters],
    queryFn: async () => {
      const res = await bookService.getCalendarEvents(filters);
      return res.data?.data ?? res.data ?? [];
    },
  });
}
