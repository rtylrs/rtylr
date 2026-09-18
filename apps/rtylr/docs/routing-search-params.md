# Route Search Params

**Status:** adopted 2026-08-10. First implementation: `paySlipViewRoute` in
`apps/hr/src/router.tsx`. Extended to path params (`params.parse`) the same
day: `payrollRunDetailRoute` in the same file.

TanStack Router routes that read query params validate them with a Zod schema
through `validateSearch`. Components read the result with `getRouteApi(...)
.useSearch()`, and callers navigate with a typed `search` prop. No route may
parse the query string by hand, and no `to=` prop may be cast to reach a route.

## Why

Before this, `apps/hr` did both halves by hand:

```tsx
// the caller: a cast, because the route was untyped
<Link to={`/payroll/my-payslips/view?runId=${run.id}` as any}>

// the route: manual parsing, unvalidated
const searchStr = useRouterState({ select: (s) => s.location.searchStr });
const sp = new URLSearchParams(searchStr);
const runId = sp.get('runId') ? Number(sp.get('runId')) : undefined;
```

The cast defeats the router's type checking entirely: rename the route or change
the param and nothing fails until a user hits a dead link. The manual parse
silently forwards nonsense. `Number('-3')` is a valid number, so the old code
issued `GET /hr/payroll/my-runs/-3`, a request for an id that cannot exist.

The root `CLAUDE.md` already requires Zod at "all external data boundaries: API
responses, form inputs, env vars, localStorage reads". A URL is the most
attacker-reachable of those, and it was the only one not covered.

## The shape

Route search schemas live with the app's other schemas, in
`src/shared/schemas/<domain>.ts`, and are re-exported from `schemas/index.ts`.
They are not a separate category of file.

```ts
export const paySlipSearchSchema = z.object({
  runId: z.number().int().positive().optional().catch(undefined),
});
```

### Never use `z.coerce`

Two independent reasons, both verified rather than assumed:

1. The router's `defaultParseSearch` already runs values through `JSON.parse`.
   `?runId=12` arrives as the number `12`, not `"12"`. Coercion solves a problem
   that does not exist.
2. `z.coerce.number()` has an **input** type of `unknown`. `<Link>`'s `search`
   prop is typed against the schema's input, not its output, so a coerced schema
   makes `search={{ runId: 'nonsense' }}` compile clean. This is the exact
   failure the schema was introduced to prevent.

Measured against the manual parsing it replaced:

| URL            | old parsing  | `z.coerce.number()` | adopted schema |
| -------------- | ------------ | ------------------- | -------------- |
| `?runId=12`    | 12           | 12                  | 12             |
| `?runId=`      | no fetch     | **0**               | no fetch       |
| `?runId=abc`   | no fetch     | **throws**          | no fetch       |
| (param absent) | no fetch     | **throws**          | no fetch       |
| `?runId=-3`    | **fetch -3** | fetch -3            | no fetch       |

### Always `.catch(undefined)` on optional params

`validateSearch` runs _before_ the route renders. A schema that throws does not
produce the feature's own empty state; it produces the router error boundary. A
stale bookmark or a hand-edited URL is not an exceptional condition, so it must
degrade to the same "not found" state the feature already renders.

Pair this with a query hook that treats `undefined` as "do not fetch"
(`enabled: !!id`), and a malformed URL costs nothing.

### `getRouteApi` takes a route id, not a path

Where a parent route is pathless, its `id` is part of every child's id. In
`apps/hr` the parent is `hrConfigRoute` with `id: 'hr-config'`, so:

- `getRouteApi('/hr-config/payroll/my-payslips/view')` (id)
- `<Link to="/payroll/my-payslips/view">` (path)

Prefer `getRouteApi` over importing the route object from `router.tsx`. The
router lazy-imports the route component, so importing it back would create a
module cycle. `getRouteApi` resolves off the `Register` interface at type level
and the router instance at runtime, with no import edge.

## Path params: the same rule, one difference

Dynamic segments (`/payroll/$id`) get the same treatment through the route's
`params` option. Reference implementation: `payrollRunDetailRoute` in
`apps/hr/src/router.tsx`, schema `payrollRunParamsSchema` in
`shared/schemas/payroll.ts`.

```ts
const payrollRunDetailRoute = createRoute({
  getParentRoute: () => hrConfigRoute,
  path: '/payroll/$id',
  params: {
    parse: (raw) => payrollRunParamsSchema.parse(raw),
    stringify: ({ id }) => ({ id: String(id) }),
  },
  // ...
});
```

The one difference from search params: the router never JSON-parses path
segments, so `params.parse` receives raw strings and the schema needs an
explicit string-to-number transform:

```ts
export const payrollRunParamsSchema = z.object({
  id: z
    .string()
    .transform((value) => Number(value))
    .pipe(z.number().int().positive())
    .optional()
    .catch(undefined),
});
```

It is still not `z.coerce`, for the same input-type reason as above - and
`Number('')` is `0`, which the `.pipe` would have to reject anyway. Everything
else carries over unchanged: `.catch(undefined)` degrades a malformed segment
to the feature's own not-found state instead of the router error boundary,
components read the parsed value with `getRouteApi('...').useParams()`, and
callers navigate with `<Link to="/payroll/$id" params={{ id }}>` - never a
template string cast with `as any`, and never `useParams({ strict: false })`
followed by a manual `Number()`.

## Verifying

`pnpm typecheck` passing does **not** prove the wiring is live. A route with no
`validateSearch`, or a schema whose input widened to `unknown`, also type-checks.
Confirm by temporarily assigning the value to a deliberately wrong type and
checking that `tsc` errors, then revert.

`e2e/hr-payslip-search.spec.ts` covers the runtime half: that a valid id resolves
the run the user clicked, that malformed ids render the empty state rather than
the error boundary, and that out-of-range ids issue no backend request at all.
`e2e/hr-payroll-run-params.spec.ts` covers the same three properties for the
`/payroll/$id` path param.
