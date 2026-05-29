export type EventType = 'Meeting' | 'Holiday' | 'Training' | 'Review' | 'Other';
export type RecurrenceRule = 'once' | 'daily' | 'weekly' | 'monthly' | 'yearly';

export function eventTypeLabel(type: EventType): string {
  const labels: Record<EventType, string> = {
    Meeting: 'Meeting',
    Holiday: 'Holiday',
    Training: 'Training',
    Review: 'Review',
    Other: 'Other',
  };
  return labels[type];
}

export function eventTypeColor(type: EventType): string {
  const colors: Record<EventType, string> = {
    Meeting: 'blue',
    Holiday: 'green',
    Training: 'purple',
    Review: 'orange',
    Other: 'gray',
  };
  return colors[type];
}

export function recurrenceLabel(rule: RecurrenceRule): string {
  const labels: Record<RecurrenceRule, string> = {
    once: 'One Time',
    daily: 'Daily',
    weekly: 'Weekly',
    monthly: 'Monthly',
    yearly: 'Yearly',
  };
  return labels[rule];
}

export function eventsOverlap(start1: string, end1: string, start2: string, end2: string): boolean {
  return new Date(start1) < new Date(end2) && new Date(end1) > new Date(start2);
}

export function isMultiDayEvent(startDateIso: string, endDateIso: string): boolean {
  const start = new Date(startDateIso);
  const end = new Date(endDateIso);
  return (
    end.getFullYear() !== start.getFullYear() ||
    end.getMonth() !== start.getMonth() ||
    end.getDate() !== start.getDate()
  );
}

export function eventDurationMinutes(startTime: string, endTime: string): number {
  const [sh, sm] = startTime.split(':').map(Number);
  const [eh, em] = endTime.split(':').map(Number);
  return eh * 60 + em - (sh * 60 + sm);
}

export function eventsOnDate<T extends { date: string }>(events: T[], dateIso: string): T[] {
  return events.filter((e) => e.date.startsWith(dateIso));
}

export function isPublicHoliday(dateIso: string, holidays: string[]): boolean {
  return holidays.includes(dateIso);
}

export function workingDaysInMonth(year: number, month: number): number {
  const daysInMonth = new Date(year, month, 0).getDate();
  let count = 0;
  for (let d = 1; d <= daysInMonth; d++) {
    const day = new Date(year, month - 1, d).getDay();
    if (day !== 0 && day !== 6) count++;
  }
  return count;
}
