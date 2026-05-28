// Pure helpers for managing order item course assignment.

import type { OrderLineItem } from '../types/pos';

type Course = NonNullable<OrderLineItem['course']>;

/** Return true if an order line item has been assigned to a course. */
export function hasCourse(item: OrderLineItem): boolean {
  return item.course !== undefined && item.course !== null;
}

/** Group order line items by their course. */
export function itemsByCourse(
  items: OrderLineItem[],
): Record<Course | 'unassigned', OrderLineItem[]> {
  const result: Record<string, OrderLineItem[]> = {
    appetizer: [],
    main: [],
    dessert: [],
    drink: [],
    other: [],
    unassigned: [],
  };
  for (const item of items) {
    const key = item.course ?? 'unassigned';
    result[key].push(item);
  }
  return result as Record<Course | 'unassigned', OrderLineItem[]>;
}

/** Return all items that belong to a specific course. */
export function filterByCourse(items: OrderLineItem[], course: Course): OrderLineItem[] {
  return items.filter((i) => i.course === course);
}

/** Return all items that have not been assigned to a course. */
export function unassignedItems(items: OrderLineItem[]): OrderLineItem[] {
  return items.filter((i) => !hasCourse(i));
}

/** Return a display label for a course. */
export function courseLabel(course: Course): string {
  const labels: Record<Course, string> = {
    appetizer: 'Appetizer',
    main: 'Main',
    dessert: 'Dessert',
    drink: 'Drink',
    other: 'Other',
  };
  return labels[course];
}

/** Return the courses present in a set of items (in display order). */
export function activeCourses(items: OrderLineItem[]): Course[] {
  const order: Course[] = ['appetizer', 'main', 'dessert', 'drink', 'other'];
  const present = new Set(items.map((i) => i.course).filter(Boolean) as Course[]);
  return order.filter((c) => present.has(c));
}
