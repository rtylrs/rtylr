import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { bookService } from '../../shared/services/bookService';
import { calendarEventSchema } from '../../shared/schemas/calendar';

export interface CalendarFilters {
  start?: string;
  end?: string;
  staffId?: number;
}

export function useCalendarEvents(filters?: CalendarFilters) {
  return useQuery({
    queryKey: ['calendar', filters],
    queryFn: async () => {
      const res = await unwrap(bookService.getCalendarEvents(filters));
      const raw = (res as Record<string, unknown>)?.events ?? res;
      return z.array(calendarEventSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
