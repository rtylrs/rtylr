import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { bookService } from '../services/bookService';

export interface CalendarFilters {
  start?: string;
  end?: string;
  staffId?: number;
}

export function useCalendarEvents(filters?: CalendarFilters) {
  return useQuery({
    queryKey: ['calendar', filters],
    queryFn: () => unwrap(bookService.getCalendarEvents(filters)),
  });
}
