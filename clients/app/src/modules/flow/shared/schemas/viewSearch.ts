import { z } from 'zod';

/**
 * `?view=<id>` selects a saved view on the Views page, so "Save as view" on a
 * board can open the view it just created. Declared on the Flow root next to
 * `openTask`, with the same string-or-number union: the router JSON-parses
 * search values, so a hand-written `?view=12` arrives as a number.
 */
export const viewSearchSchema = z.object({
  view: z
    .union([z.string(), z.number()])
    .transform((value) => String(value))
    .optional()
    .catch(undefined),
});
