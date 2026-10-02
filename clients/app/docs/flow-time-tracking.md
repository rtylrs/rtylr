# Flow time tracking: what shipped, and what the contract cannot hold

Batch 1 of the Flow time-tracking design: task-detail hours, workspace time
settings, and board customization. Three of fifteen designed screens. This
document records the pattern the batch introduced and the contract gaps that
shaped what the screens actually contain, so the next person does not rediscover
them from a rendered mock.

## The pattern: a settings store as the endpoint seam

`apps/flow/src/shared/stores/timeSettingsStore.ts`

Four designed screens read the same workspace values (week start, rounding
increment, weekly target, overtime threshold). None of them can read those values
from the server, because there is no workspace time-settings endpoint.

The alternative to a seam is four screens each hardcoding `7` and `40` at their
own point of use, which is how a workspace setting becomes twelve constants. The
store holds the shipped defaults in one place. Landing the endpoint replaces this
store's internals: the hook signature, the field names and every call site stay
as they are.

Two rules that make it a seam rather than a lie:

1. **The UI states the limitation.** `TimeSection` opens with an `Alert` saying
   the values are saved in this browser only. A settings screen that silently
   fails to be a workspace setting is worse than no settings screen.
2. **Nothing derives a server-side consequence from it.** The store never decides
   what an endpoint would have decided. `roundLoggedMinutes` shapes a value
   before it is posted, which is a client concern; nothing reads
   `approvalLocksEntry` to decide whether an entry is editable, because the
   server has no approval state to be consistent with.

Reach for this pattern when a designed surface needs values that have a clear
owner and no endpoint yet. Do not reach for it when the value is already
server-owned, or when a wrong local value would produce a wrong write.

## Contract gaps, measured against `packages/voxie/api/flow/task/v1/models.onk`

The design doc labels screen 1 "BACKEND COMPLETE" and screen 11 "ALL ENDPOINTS
EXIST". Both are true about the _endpoints_ and false about the _fields_. What
the contract actually carries:

```
message TimeLog { id, task_id, user_id, duration_minutes, description, logged_at, created_at }
CreateTimeLogRequest { duration_minutes @gt(0), description @required, task_id }
```

Three RPCs: create, list by task, delete by id.

| Designed                                                                          | Contract reality                                | What shipped instead                                                                       |
| --------------------------------------------------------------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Status pills (approved / pending / locked) and the per-status split in the header | no `status` field                               | absent; a pill the server never sends is invented data                                     |
| `Billable` toggle on the form, `Billable` pill on rows                            | no `billable` field                             | absent                                                                                     |
| Category chips, and `· Development ·` on each row                                 | no `category` field                             | absent from the log form; the settings list carries a `GateNotice` saying nothing reads it |
| Category "Logged" column, delete-with-reassign dialog                             | no per-category hours exist                     | absent; there is nothing to reassign                                                       |
| Date picker on the log form                                                       | `CreateTimeLogRequest` has no date              | present, fixed to today, disabled, with the reason stated                                  |
| Editing an entry                                                                  | no PATCH                                        | not offered; delete and re-log is the correction path                                      |
| `description` optional                                                            | `@required` in the contract                     | `flowApi` already substitutes `'-'`; `TimePanel` treats `'-'` as empty on read             |
| WIP limit per column                                                              | `ApiProjectColumn` is `name, color, sort_order` | present, disabled, reason names the missing field                                          |
| "Prompts for hours" column flag                                                   | same                                            | present, disabled, reason stated                                                           |
| Four card-field toggles                                                           | no project or column card-field set             | present, disabled, reason stated                                                           |

The rule applied throughout: **a control whose value cannot be stored is either
absent or disabled with the reason beside it, never drawn as though it works.**
`GateNotice` exists so that reason is one component rather than eleven phrasings.

The distinction between the two treatments is whether the _concept_ is real. A
WIP limit is a real board concept with nowhere to persist, so it is shown
disabled. An entry status is not a concept the system has at all, so drawing one
would teach the user something untrue.

## Deviations from the design, deliberate

- **Reorder is buttons plus `Alt`+arrow, not drag.** The board's existing drag
  handling is raw HTML5 events with no touch support. Up/down buttons serve
  pointer, touch and keyboard from one implementation, and the design already
  required them below 768px and for keyboard users.
- **Dates follow the locale's own day/month order.** The mocks show `Tue 11 Aug`
  throughout, which is British ordering. `formatEntryTimestamp` uses `Intl` and
  lets `en-US` render `Tue, Aug 11`. Digits are pinned to Latin with the
  `nu-latn` extension, because a duration read into a payroll import must not
  arrive in Eastern Arabic numerals.

## Token changes

`flow-status-tokens.css` from the design project is applied in
`apps/flow/src/globals.css`. Additions are inert until used. Two edits changed
existing values app-wide and were signed off separately:

- `--flow-text-muted` `.5` → `.58` dark, `.5` → `.62` light. At `.5` it measured
  4.39:1 on `--flow-card` and 3.74:1 on white, both failing AA for the 10–12px
  type this feature is built from. Now 5.99:1 and 5.75:1. This also fixes
  `SidebarLink`'s badge, which measured 2.14:1.
- `--accent-muted` gains a light-mode value of `.18`, since a `.12` wash reads
  lighter on white than on black.

`--flow-text-dim` keeps its value and narrows its use: rules, dividers, and icon
strokes that repeat an adjacent label. It never carries text: 2.24:1 cannot be
fixed by nudging alpha without becoming muted. `DetailField` in `TaskDetail.tsx`
moved its label off it.

**Every alias is redeclared in the light block.** A `var()` resolves where it is
declared, so `--flow-status-pending: var(--flow-warn-fg)` declared only in
`:root` paints the dark tone in light mode. This is the most common failure mode
in a themed token set and it is silent.

## Two defects live verification found that the test suite could not

Both were invisible to `typecheck`, `eslint` and 997 passing unit tests. Both were
found by driving the real app in a browser with the API stubbed.

### `--accent-muted` was overridden by an inline style

`FlowLayout` applies `--accent`, `--accent-muted`, `--accent-border` and
`--accent-focus` as **inline styles on `documentElement`**, derived from the active
workspace colour. Inline styles beat every stylesheet rule, so the signed-off
light-mode `--accent-muted: .18` in `globals.css` was dead on every authenticated
screen - the layout re-applied `.12` the moment it mounted.

The per-mode alpha now lives in `accentTints` (`shared/lib/colorUtils.ts`), which
`FlowLayout` calls with the resolved theme, so it follows both the mode and the
per-workspace accent. Measured through the real theme toggle: dark `.12`/`.3`,
light `.18`/`.42`.

**The general rule:** a custom property that any code sets via `style.setProperty`
cannot be themed from a stylesheet. Check for an inline writer before adding a
per-mode value to `globals.css`.

### The shared preset never reached any app, so `bg-accent` was transparent

Every `globals.css` in this repo opened with a comment saying the
`@voxire/tailwind-config` extensions were "ported below into v4's native `@theme`
block". **None had been.** The cause was not a forgotten paste: the preset is a
**v3-style JS config, which Tailwind v4 does not read**, and the apps
deliberately avoid `@config` because it breaks v4's content scanner here. So the
whole preset was inert. Measured on real rendered elements, `bg-accent` and
`bg-warning` computed to `rgba(0,0,0,0)`.

It was never only colours. Also dead: `animate-spin-slow` (so `Button`'s loading
spinner did not spin), `shadow-modal` (so `Dialog` had no shadow), and the radius
scale.

Fixed by porting the preset to `packages/tailwind-config/theme.css` and importing
it from each app after `@import 'tailwindcss'`. One definition, ten one-line
imports. Each app re-points individual tokens at its own runtime custom
properties from its own `:root`, which is how Flow's `bg-accent` follows the
per-workspace teal instead of the preset's cyan:

```css
:root {
  --color-accent: var(--accent);
  --color-accent-foreground: var(--flow-on-accent);
  --color-warning: var(--flow-warn-fg);
}
```

Flow re-points warn and danger too, so a shared `Alert` renders in Flow's palette
with no per-instance override. `text-warning` measures `#8f5f14` in light mode.

Measured blast radius before applying: 45 direct utility usages across all ten
apps, 42 of them in dash, and 13 `@voxire/ui` components rendered in total, of
which **zero** were `Button` or `Badge`.

**A caution on verifying this.** Tailwind v4 compiles only what it finds by
scanning source. A probe that injects a class at runtime measures a class that
was never generated, and reports every token as broken even when it works: that
first probe had even `rounded-lg` reading `0px`. Measure real rendered elements.

## Changes made to `@voxire/ui`

All four landed in the packages submodule rather than being forked locally:

- **`Button` accent foreground** is now `text-accent-foreground`, near-black.
  `bg-accent text-white` measured 1.9:1 and failed AA at every size. This one is
  coupled to the theme block above: while `bg-accent` was transparent the white
  text was legible only by accident, so landing the theme block on its own would
  have _introduced_ the failure. They ship together, or the fix goes first.
- **`Badge` `sm` at 10px**, so a pill beside 11px body text does not force a 36px
  row taller.
- **`Button` `xs` at 30px**, for toolbar and dense-row actions.
- **`Avatar` `xs` at 24px/10px**, for the 32–44px rows on the later screens.

The last three are additive and every default is unchanged, so nothing existing
moved. Worth knowing when reviewing: across all ten apps there are currently
**zero** `Button` and `Badge` instances rendered from `@voxire/ui`, so the
contrast fix changes no button on screen today. It is correctness for the next
one.

`RealtimeClient` in `@voxire/api` also gained `getStatus` and `subscribeStatus`
over four states, since it previously exposed only `subscribe` and had an empty
`onerror`. `reconnecting` is deliberately distinct from `disconnected`: the first
retries with backoff and writes still work, the second needs a reload. Flow's
`useRealtimeStatus` wraps it and `TimePanel` renders the warning.

### The mobile task detail hid its own title, and ordering was not why

At 390px the task title and description were not on screen. The obvious reading
was that the metadata rail stacked above them, and the obvious fix was to reorder.
Both were wrong: the rail is **already last in the DOM**, so `order-last` was a
no-op that changed nothing.

The real cause is that the main column carried `flex-1`, whose `flex-basis` is
`0%`. Inside a fixed-height scrolling flex column the taller rail claimed every
pixel, and the main column collapsed to nothing. It now takes its natural height
below `lg` and shares the row only from `lg` up.

It also needs `min-w-0`: the title is an `<input>`, and once the column's basis
was `auto` the input's intrinsic width pushed the column past the viewport.
Verified at 390px that `scrollWidth === clientWidth` on task detail, settings and
board, so nothing overflows sideways.

**The lesson worth keeping:** a symptom that looks like source order can be a
`flex-basis` collapse. Check computed geometry before rearranging markup, or you
ship a no-op and believe it worked.

### Settings sections are linkable

The `Time` sidebar group needed a destination, and settings sections lived in
`useState`. `/settings` now takes a validated `?section=` search param, so any
section can be linked. Zod enum, **no `z.coerce`** (the router already JSON-parses
query values, and coerce's `unknown` input type would let a wrong-typed `<Link>`
compile), and `.catch(undefined)` so a stale URL does not replace the screen with
the router error boundary. `SidebarLink` gained an optional `search` prop.

`Time` is a fourth id in `SIDEBAR_SECTION_IDS` and inherits collapse, reorder and
persistence. No storage migration was needed: `normalize()` drops unknown ids and
appends known ids missing from a stored order, so an older stored order simply
gains the group at the end. Settings → Time is its only entry until the blocked
screens land.

## Build order for what remains

Twelve screens are still blocked, on four things. In the order that unblocks the
most:

1. **`PATCH /time-logs/{id}` and a `date` on create.** Timesheet cannot exist
   without both, and the date field on screen 1 stops being disabled.
2. **A grouped aggregation endpoint.** Reports, Portfolios, Goals, the project
   Time tab, and the logged-hours half of Workload all read it.
3. **Entry status and approval state.** Review, and the status pills across every
   screen that renders an entry.
4. **`category` and `billable` on `TimeLog`**, plus a rates table. Until then
   Settings → Time configures categories nothing reads.

`StatusPill`, `TimesheetCell`, `WeekGrid`, `PivotTable`, `CapacityCell` and
`BulkActionBar` are named in the design's component inventory and are not built:
each has zero call sites until the screens above exist.
