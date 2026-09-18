import { z } from 'zod';

/**
 * `?openTask=<id>` opens task detail as a drawer over whatever list is showing,
 * instead of replacing it (rtylr#1079).
 *
 * Declared on the **root** route so it is valid on every screen a task can be
 * opened from: the board, the list, both calendars, My tasks, Today, Triage,
 * Archive, Team and the command bar. The alternative was a `validateSearch`
 * entry on each of 38 routes.
 *
 * Named `openTask` rather than `task` because `/time` already has a `task`
 * param meaning "preselect this task in the Add hours form"
 * (`myDaySearchSchema`). One name, two meanings, on a route you can reach from
 * a task, is a collision waiting to happen.
 *
 * ## Why the union
 *
 * The router's `defaultParseSearch` runs every value through `JSON.parse`, so
 * `?openTask=101` arrives as the **number** `101` while a task id is a string
 * everywhere else in the app. A plain `z.string()` would reject every numeric
 * id and `.catch(undefined)` would silently swallow it, leaving a drawer that
 * never opens and no error to explain why. The union accepts what the router
 * actually delivers and normalises to the string the rest of the code uses.
 *
 * Not `z.coerce.string()`: its input type is `unknown`, which would let a
 * wrong-typed `search` prop compile clean. See docs/routing-search-params.md.
 *
 * ## What the round trip actually looks like
 *
 * The router JSON-*stringifies* on the way out as well, so navigating with the
 * string `'101'` produces `?openTask=%22101%22` and parses back as the string
 * `'101'`. Ugly to read, but exact: an int64 id keeps every digit, because it
 * never becomes a double.
 *
 * The number branch of the union therefore only serves a **hand-written** URL,
 * `?openTask=101` without the quotes, which is the one case where an id above
 * 2^53 - 1 would lose precision before any schema sees it. Not guarded, because
 * nothing here could recover the lost digits; recorded so the union's purpose
 * is not mistaken for the app's own behaviour.
 */
export const taskDrawerSearchSchema = z.object({
  openTask: z
    .union([z.string(), z.number()])
    .transform((value) => String(value))
    .optional()
    .catch(undefined),
});

export type TaskDrawerSearch = z.infer<typeof taskDrawerSearchSchema>;
