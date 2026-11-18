import { z } from 'zod';
import { parseTaskFilters, serializeTaskFilters } from '../filters/taskFilters';

/**
 * `?filter={...}` carries the task filters of whatever list is showing, so a
 * filtered board can be shared as a link, the way Jira puts its JQL in the URL.
 *
 * Declared on the Flow root route next to `openTask` for the same reason: it is
 * valid on the board, the list, the calendar, My tasks and Team tasks, and one
 * entry beats five.
 *
 * The router JSON-encodes object values, so the param arrives as an object.
 * Anything malformed in it (a hand-edited link, a field this build does not
 * know) is dropped value by value in `parseTaskFilters` rather than throwing
 * the whole filter away.
 */
export const taskFilterSearchSchema = z.object({
  filter: z
    .record(z.string(), z.unknown())
    .optional()
    .catch(undefined)
    .transform((value) =>
      value === undefined ? undefined : serializeTaskFilters(parseTaskFilters(value))
    ),
});
