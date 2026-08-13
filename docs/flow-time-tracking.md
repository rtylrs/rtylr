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
| Status pills (approved / pending / locked) and the per-status split in the header | no `status` field                               | absent; a pill the server never sends is invented data                                    |
| `Billable` toggle on the form, `Billable` pill on rows                            | no `billable` field                             | absent                                                                                     |
| Category chips, and `· Development ·` on each row                                 | no `category` field                             | absent from the log form; the settings list carries a `GateNotice` saying nothing reads it |
| Category "Logged" column, delete-with-reassign dialog                             | no per-category hours exist                     | absent; there is nothing to reassign                                                      |
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

- **No socket-down indicator.** `RealtimeClient` in `@voxire/api` exposes only
  `subscribe`; there is no connection state to read and its `onerror` handler is
  empty. A connection indicator is a change to that shared package, which nine
  apps consume.
- **Reorder is buttons plus `Alt`+arrow, not drag.** The board's existing drag
  handling is raw HTML5 events with no touch support. Up/down buttons serve
  pointer, touch and keyboard from one implementation, and the design already
  required them below 768px and for keyboard users.
- **Dates follow the locale's own day/month order.** The mocks show `Tue 11 Aug`
  throughout, which is British ordering. `formatEntryTimestamp` uses `Intl` and
  lets `en-US` render `Tue, Aug 11`. Digits are pinned to Latin with the
  `nu-latn` extension, because a duration read into a payroll import must not
  arrive in Eastern Arabic numerals.
- **No `Time` sidebar section.** `SIDEBAR_SECTION_IDS` drives hardcoded per-id
  blocks in `FlowLayout.tsx`, and every Time destination (Timesheet, Review,
  Reports, Workload) is blocked. Adding the id now ships a section header with
  nothing under it.
- **The mobile task-detail rail is above the description, not below it.** The
  design has it stacking underneath. `TaskDetail`'s own layout puts the rail first
  at narrow widths, and reordering it would move all twelve rail fields, not just
  Time. Left alone deliberately; worth fixing as its own change.

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

### No app defines a Tailwind `@theme` block, so `bg-accent` is transparent

Every `globals.css` in this repo opens with a comment saying the
`@voxire/tailwind-config` extensions are "ported below into v4's native `@theme`
block". **No app actually has one.** Measured in the running app, `bg-accent` and
`bg-warning` both compute to `rgba(0,0,0,0)`.

Every `@voxire/ui` component that colours itself through those utilities is
therefore untinted app-wide:

| Component | Class                                          | Effect                                      |
| --------- | ---------------------------------------------- | ------------------------------------------- |
| `Switch`  | `data-[state=checked]:bg-accent`               | checked is indistinguishable from unchecked |
| `Alert`   | `bg-warning/10 border-warning/30 text-warning` | plain bordered box, no tone                 |
| `Avatar`  | `bg-accent/20 text-accent`                     | untinted circle                             |
| `Badge`   | `bg-{tone}/10 text-{tone}`                     | invisible tint                              |
| `Button`  | `accent: bg-accent text-white`                 | no fill                                     |

This is pre-existing and much larger than this feature. `FlowSwitch`
(`shared/components/FlowSwitch.tsx`) is a **local workaround, not the fix**: it
passes arbitrary-value utilities pointing at Flow's own custom properties, which
resolve without a theme block. The real fix is a theme mapping in the consuming
apps, and it would change the appearance of every existing use of these utilities
across ten apps - which is a sign-off, not a patch.

## Outstanding requests to `@voxire/ui`

Written down rather than forked locally. Each affects every app using the
component:

- **`Button` accent foreground.** The variant is `bg-accent text-white`, which
  measures 2.05:1 and fails AA at every size. `--flow-on-accent` (`#0a0a0a`,
  9.6:1) is what Flow uses at its own call sites; the component still ships
  white.
- **`Badge` `sm` at 10px.** `text-xs` badges force a 36px table row taller.
- **`Button` `xs` at 30px**, for toolbar and dense-row actions.
- **`Avatar` `xs` at 24px/10px**, for the 32–44px rows on the later screens.

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
