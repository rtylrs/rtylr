import { z } from 'zod';

/**
 * `/time` search params. A task's own "Add hours" action can route here with
 * a date, project and task preselected instead of opening a second form.
 *
 * No `z.coerce`: the router already JSON-parses query values, and coerce's
 * input type is `unknown`, which would let a wrong-typed `<Link search={...}>`
 * compile clean. `.catch(undefined)` so a stale or hand-edited URL falls back
 * to the form's own defaults instead of replacing the screen with the router
 * error boundary. See docs/routing-search-params.md.
 */
export const myDaySearchSchema = z.object({
  date: z.string().optional().catch(undefined),
  project: z.string().optional().catch(undefined),
  task: z.string().optional().catch(undefined),
});
