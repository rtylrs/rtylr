# Flow Appearance: enriched personalization

Date: 2026-08-22
Status: approved for planning
Scope: `packages/utils`, `packages/ui`, `apps/flow`

## 1. Why

Flow's Appearance settings section is 45 lines: a heading, a label, and three
theme cards (`apps/flow/src/features/settings/FlowSettings.tsx:455`). The same
switcher is duplicated inside `apps/flow/src/layouts/FlowLayout.tsx:1230`, so the
dedicated settings page is the weaker of the two entry points.

The section is thin for a structural reason, not through neglect. Flow's accent
is derived from the **workspace** colour, never from the user, so Appearance had
nothing user-owned to offer beyond light/dark. Adding user-owned appearance means
deciding what a personal accent does when it disagrees with a workspace accent,
and deriving the foreground tokens that a free-form colour choice implies.

## 2. What was measured, and what it corrected

An interactive prototype was built before this spec and used to measure the
existing token behaviour rather than reason about it. Two hypotheses held during
design were falsified by it. Both are recorded because both changed the design.

### 2.1 `--accent-fg` does not fail contrast. It drifts in hue.

`--accent-fg` is hardcoded per mode in `apps/flow/src/globals.css:103` (`#64c5b9`
dark, `#1f7d72` light) while `--accent` follows the active workspace colour. The
initial assumption was that a user-chosen accent would therefore ship a WCAG
failure.

Measured across 8 palettes and 14 accents, the hardcoded teal holds 8.81:1 on
`--flow-card` in dark mode and roughly 4.9:1 on white in light mode, for every
accent. It never fails AA, because it was chosen to clear the bar on those two
grounds and it keeps doing so no matter what sits beside it.

The real defect in this token is hue drift: with a lime accent the foreground
sits about 90 degrees off the fill next to it. **This is a cohesion defect, not
an accessibility one**, and it ships today. Selecting the violet or red workspace
colour already leaves chip text teal.

### 2.2 `--flow-on-accent` is the token that actually breaks.

`--flow-on-accent` is hardcoded near-black (`#0a0a0a`). On a dark accent the text
sitting on the accent fill measures between 1.30:1 and 1.63:1. That is
unreadable, not marginal.

It is unreachable from a curated palette. All ten proposed swatches, and all
eight existing `WORKSPACE_COLORS`, are light enough that near-black stays safe.
The tightest existing value is `#8b5cf6` at 4.68:1, which passes. **There is no
live accessibility bug today.** The failure becomes reachable only when a
free-form colour picker is introduced, which section 3 confirms it will be.

### 2.3 A latent value to leave alone

`#6366f1` in `PROJECT_COLORS` (`apps/flow/src/shared/lib/colorUtils.ts:48`)
measures 4.43:1 against near-black. It is latent rather than live because project
colours paint 7px dots that carry no text and never become `--accent`. Do not
route `PROJECT_COLORS` into the accent path without deriving the foreground.

### 2.4 A margin the derivation needs

Stopping the derivation at the first value that reaches 4.5:1 produced exactly
4.50:1 on every light palette, leaving nothing for rounding or a later change to
a surface token. Targeting 4.6:1 costs an imperceptible darkening. Verified
across 8 palettes times 14 accents: 112 combinations, zero AA failures, worst
case 4.60:1.

## 3. Decisions

| Decision          | Choice                                        | Rationale                                                                                                                                                                         |
| ----------------- | --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Ambition          | Expressive personalization                    | Presets, accent choice, density, text size, corners, motion. Not a token-only panel, not a full theme studio.                                                                     |
| Where it lives    | Extract to `packages`, Flow is first consumer | `packages/utils/src/hooks/useTheme.ts` already exports `createThemeStore`, and Flow, dash, and menu each hand-rolled their own instead. Four divergent copies is the alternative. |
| Accent precedence | User accent overrides workspace               | A personal choice beats an org default. Workspace colour still paints the workspace's own dot and chip, so the identity cue survives.                                             |
| Accent input      | Curated swatches **and** custom hex           | Makes the derivation mandatory rather than optional. This is the decision that selects the architecture in section 4.                                                             |
| Panel scope       | All seven controls, one of them disabled      | Corners and density turned out cheap (section 7). Text size cannot work against px literals, so it ships disabled with the reason stated rather than silently absent.             |
| Persistence       | Per-device, stated in the UI                  | No appearance contract exists in `packages/voxie/api`. Nothing server-side reads these values, so per-device is honest rather than a broken setting.                              |

## 4. Architecture

Three pieces, each independently testable. The shape is chosen so the contrast
derivation is a pure function, which is the only way to prove the accessibility
claim instead of asserting it.

### 4.1 `createAppearanceStore` (packages/utils)

Pure state. No DOM writes.

```ts
interface AppearancePrefs {
  mode: 'light' | 'dark' | 'system';
  palette: string; // preset id within the resolved mode's family
  accent: string | null; // null means "follow the workspace colour"
  density: 'compact' | 'cozy' | 'roomy';
  textSize: 's' | 'm' | 'l';
  corners: 'sharp' | 'soft' | 'round';
  reduceMotion: boolean;
}
```

`accent: null` is load-bearing. It is what makes "user accent overrides
workspace" expressible without storing a copy of the workspace colour, and it is
what a reset control writes.

Two constraints on the extraction:

1. **Flow's behaviour moves into the package, not the reverse.** The existing
   shared factory never sets the `data-theme` attribute; only Flow's hand-rolled
   copy does. Adopting the shared factory as written would silently break every
   `:root[data-theme='light']` rule in `globals.css`.
2. **Migrate the legacy key.** Read the existing `flow-theme` localStorage entry
   on init so anyone who already chose light keeps it.

### 4.2 `appearanceTokens(prefs, ctx)` (packages/utils)

Pure function returning `Record<string, string>`. No DOM, no React.

`accentTints` moves here from `apps/flow/src/shared/lib/colorUtils.ts:86` and
gains the two members it should always have had:

- `--accent-fg`: walk the accent's own HSL lightness until it clears **4.6:1** on
  `--flow-card`. Hue and saturation are preserved, so the result still reads as
  the colour the user chose. Fall back to black or white if the walk hits its
  bound.
- `--flow-on-accent`: choose between near-black and white by measuring both
  against the accent. Never assume.

It also emits `--flow-u` (density), `--flow-fs` (text size), and `--flow-r`
(corners) so section 7 has something to consume.

### 4.3 `<AppearanceProvider>` (packages/ui)

The **single** writer. Reads the store, calls the derivation, applies the result
to `documentElement` in one effect. Paired with an `initAppearance()` called from
`main.tsx` before render so there is no flash of the wrong palette.

`FlowLayout` stops writing accent tints (`FlowLayout.tsx:498`). The workspace
colour becomes an input to the provider as the fallback when `accent` is null,
which is what accent precedence means in practice.

This retires a documented landmine rather than working around it. An inline
custom property on `documentElement` outranks every stylesheet rule, so today
`globals.css` cannot theme anything `FlowLayout` writes. One writer removes the
conflict instead of adding a second party to it.

## 5. The panel

Rewrite of `AppearanceSection`. Seven controls, in this order:

1. **Mode**: light / dark / system, as miniature renders rather than flat colour
   swatches. The current cards paint `#f4f4f5` and `#0a0a0a` rectangles, which is
   the least informative preview available.
2. **Palette**: named presets within the resolved mode's family, four per mode so
   the grid is balanced in both. Each card is a miniature render. A one-line
   description per preset states what it is for.
3. **Accent**: ten curated swatches plus a hex field and a native colour input. A
   reset control writes `accent: null` and returns to the workspace colour.
4. **Density**: compact / cozy / roomy.
5. **Text size**: small / default / large.
6. **Corners**: sharp / soft / round.
7. **Reduce motion**: defaults on when the system asks for it.

**A preview pane, after all.** The first draft of this spec argued against one:
the panel lives inside Flow, so the application is already the preview. That
reasoning was incomplete. A preview lets a user **audition** a palette without
committing the whole app to it, which is the difference between trying six
palettes and being willing to try one. It also gives density and corners
somewhere to demonstrate themselves on more than one shape.

**Audition on hover, commit on click. No Apply button.** A draft-plus-Apply model
was the obvious way to build this and it is the wrong one: it contradicts
"applies instantly, nothing to save", and it introduces two-tier state whose
failure mode is showing the user a palette they are not actually getting.

Instead, the preview pane reflects **applied** state by default, and shows a
**hovered or focused** option instead whenever one is active. Point at a palette
card and the pane shows that palette; move away and it snaps back; click and the
whole app changes. Auditioning costs a mouse movement, committing stays instant,
and there is exactly one source of truth.

Requirements, so it stays cheap:

- Renders the same token consumers as the app at miniature scale: a rail, a task
  row with chip and avatar, a primary button, a progress bar.
- Takes an optional `previewOverride: Partial<AppearancePrefs>` prop. The panel
  sets it from `onPointerEnter` / `onFocus` and clears it on leave and blur. No
  store writes, no second copy of the preferences.
- Derives its tokens by calling the same `appearanceTokens()` the provider uses,
  and applies them **scoped to its own subtree**, never to `documentElement`.
  This is the one place a second token writer is legitimate, and it is legitimate
  precisely because it writes to a container rather than the root.
- Local component under `features/settings/components/`. Flow's panel, not a
  shared primitive, until a second app needs one.

**Touch has no hover.** On touch, the pane simply shows applied state and
auditioning happens by tapping, which is safe because every change is instant and
reversible. Do not build a long-press affordance for this.

**De-duplication.** `FlowLayout`'s inline switcher keeps its quick mode toggle,
because it is a genuinely useful shortcut, but both call sites render one shared
`<ModeCards>` component reading one store. Two implementations of the same
control become one.

**The per-device statement.** The panel says these settings apply to this device.
Required by the settings-store-as-a-seam rule: a settings screen that silently
fails to be a workspace setting is worse than no settings screen.

## 6. Responsive, i18n, and a11y

Requirements, not polish. Each item below is a rule the prototype broke at least
once before it was fixed.

- **Mobile first, `min-width` only.** Panel is single column at 390px. The accent
  grid uses capped columns, not `1fr`: at full panel width `1fr` stretched each
  swatch to 178px, which reads as a colour bar rather than a swatch.
- **Two-column layouts need roughly 1140px, not 980px.** At 980px a 392px panel
  leaves the content column 530px, which truncated every task title to an
  ellipsis.
- **44px hit areas.** The switch reads as a 26px track inside a 44px control.
  Sizing the control to the track fails the floor.
- **16px minimum on the hex input**, or iOS auto-zooms on focus.
- **Reduced motion must cover `::after` as well as `::before`.** The switch knob
  is an `::after`. Omitting it means the setting still animates the control it is
  most likely to be toggled from.
- **RTL** via logical properties, matching the repo's existing pattern.
- **i18n**: roughly 25 new keys across en, ar, and fr. Eight palette names and
  ten accent names are user-facing strings and cannot be hardcoded.
- Keyboard navigable throughout, with a visible focus state on every control.

## 7. Density, corners, and text size: the premise was wrong

The first version of this section assumed literal Tailwind classes cannot
respond to a custom property, and therefore that honouring these controls meant
migrating hundreds of call sites. **That assumption was false.** A survey
measured it and a compiled probe confirmed it independently.

Tailwind v4, as this repo configures it, already routes its utilities through
theme custom properties:

```css
.p-3 {
  padding: calc(var(--spacing) * 3);
}
.gap-2 {
  gap: calc(var(--spacing) * 2);
}
.h-10 {
  height: calc(var(--spacing) * 10);
}
.rounded-lg {
  border-radius: var(--radius-lg);
}
.text-xs {
  font-size: var(--text-xs);
}
.rounded {
  border-radius: 0.25rem;
} /* the one exception */
```

So overriding `--spacing` in `:root` moves every spacing utility in the app at
once. This is not a new technique here: `apps/flow/src/globals.css:177` already
re-points `--color-accent` at a runtime property by exactly this mechanism.

### Verdicts, which differ per control

| Control   | Verdict                                   | The number behind it                                                                                                                                                                                                         |
| --------- | ----------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Corners   | **Ship first.** Cheapest.                 | 170 of 335 addressable radius declarations already resolve through `var(--radius-*)`. The remaining 165 are bare `rounded`, fixed by one rename to `rounded-sm`, which has zero existing uses in Flow and so cannot collide. |
| Density   | **Ship second.** Roughly one line of CSS. | 1368 of 1409 spacing declarations (97.1%) already resolve through `var(--spacing)`.                                                                                                                                          |
| Text size | **Cannot work as designed.**              | 618 of 637 font-size declarations (97.0%) are px literals like `text-[12px]`. No token reaches them.                                                                                                                         |

**The sequencing in the earlier draft was backwards.** Corners goes first, not
last: it is the cheapest, and it exercises the provider's token plumbing across
170 declarations without touching a single component.

### The four-category scope is abandoned

The earlier draft proposed migrating only list rows, cards, chips, and nav items.
That is both unnecessary, given the above, and would not have worked: those
components account for 50 spacing declarations, 3.5% of the app's 1409. The
shared components are also not used by the screens the draft named. `TaskRow` is
imported only by Views and Triage; `Inbox.tsx` imports none of them and hand
rolls its own rows. Measured against each screen's full render tree, that scope
would have moved 16% of Inbox and 27% of My Tasks. Well under half, exactly the
misleading-control problem it was meant to avoid.

### Text size: disabled, with the reason stated

The concept is real. Text scaling is a genuine accessibility need, not a
decoration, so leaving it out entirely teaches the user nothing. But shipping a
control that moves 3% of the app's type would be a lie.

Per the repo's own rule for a control that cannot do its job, it ships
**disabled with the reason beside it**, using `GateNotice`. The reason is
specific and true: Flow's type sizes are px literals, so a multiplier cannot
reach them yet.

Making it work later is a bounded, mechanical migration: 618 declarations from
`text-[Npx]` to Tailwind's `text-*` scale, which resolves through `var(--text-*)`.
It is not attempted here because the scale will not match every arbitrary px
value exactly, so it changes rendering and needs its own visual review.

### Two hazards before a density swap

1. **`--spacing` also backs 447 dimension declarations** (`h-*` 157, `w-*` 136,
   `min-h-*` 58, `min-w-*` 44, and others), so changing it moves 1856
   declarations, not 1409. Mostly desirable, but 19 large layout dimensions
   (`w-64`, `w-48`, `h-40`) should be pinned to fixed values first, or the
   sidebar and panels resize with density.
2. **53 fixed heights already sit under 44px at multiplier 1.0.** Compact density
   makes those worse, so the touch-target floor has to be checked at the
   smallest multiplier, not the default.

### A pre-existing defect found on the way

`globals.css:98` states Flow's type floor is 10px. Eleven declarations already
breach it at multiplier 1.0: ten `text-[9px]` and one `text-[9.5px]`. That is a
live accessibility defect independent of this feature, and it is tracked
separately rather than fixed here.

## 8. Testing

### The contrast sweep, and why it is not sufficient on its own

Every palette times every accent fixture, asserting `--accent-fg` and
`--flow-on-accent` both clear 4.5:1. The fixture list is the ten curated swatches
**plus four deliberately dark hexes** (a navy, a plum, a forest, a brick), which
is what makes it 8 times 14 rather than 8 times 10. The dark four are the only
values that exercise the `--flow-on-accent` failure at all, so a sweep over
curated swatches alone would pass while proving nothing.

Measured result: 112 combinations, zero failures. Worst `--accent-fg` is 4.6010:1
on the linen card with the iris accent. Worst `--flow-on-accent` is 6.6117:1.

**The sweep alone cannot detect a dead derivation gate, and this was proven, not
assumed.** Forcing the `>= CONTRAST_TARGET` comparison to always-false makes the
lightness walk exhaust to pure black or white. Black on a dark card measures
16.19:1, so every floor assertion in the sweep still passes over completely dead
logic.

What catches it is the **hue-preservation test**: with the gate dead, drift goes
from about 1 degree to 172.6 degrees. That test is therefore load-bearing rather
than a nicety, and it must not be deleted as redundant. Anyone tempted to trim
the suite should read this paragraph first.

### The oracle idea does not survive contact

An earlier draft proposed asserting the derivation reproduces Flow's
hand-authored pairs within tolerance. That is not achievable, and the reason is
instructive: the shipped light foreground `#1f7d72` measures 4.9590:1 on white,
while a first-crossing walk at target 4.6 stops one stop brighter at `#308177`
and 4.6336:1. Both clear AA and both are the same hue. Asserting equality would
assert that a hand-picked value and a first-crossing walk agree exactly, which
they cannot.

Assert **hue proximity** (within 6 degrees) and **floor clearance** instead.

### The rest

- **Exhaustion is reachable and must be tested live.** There exist surfaces where
  neither pure black nor pure white clears 4.6: `#006cff` measures 4.589 against
  black and 4.576 against white. Use it as a fixture so the fallback branch is
  proven rather than dead code.
- **Mutation probes on every gate, both directions.** Required by this repo's
  convention. A probe that stops detecting a mutation is not automatically a
  regression: when the provider started importing the pure `resolveMode`, one
  probe went from 4 failures to 2 because an errant subscription can no longer
  corrupt an explicit mode's resolution. The mutation became genuinely harmless.
  Record why sensitivity changed rather than restoring it artificially.
- **Store migration.** A stored `flow-theme` value survives the move, and legacy
  dark must beat an explicitly light default, or the test passes for the wrong
  reason because dark is also the default.
- **Provider applies before paint.** No flash of the default palette on load.
- **`accent: null` falls back to the workspace colour**, and a stored accent
  overrides it.
- **One integration test with no mocks.** The unit suites mock across the package
  boundary so each side stays independently green, which means nothing proves the
  two halves meet. That check belongs at the Flow wiring site, where a real store
  exists.

## 9. Out of scope

- **dash and menu** stay on their hand-rolled stores. They can adopt the package
  later; forcing them into this change inflates the review surface for no gain.
- **Cross-device sync.** Needs a contract in `packages/voxie/api` that does not
  exist. Appearance is per-device in Linear, VS Code, and Figma by design, so
  this is a deliberate boundary rather than a gap to apologise for.
- **Per-workspace accent override.** Considered and not chosen.
- **Custom wallpapers, per-token overrides, shareable themes.** The full
  customization studio was rejected in favour of expressive personalization.

## 10. References

- Prototype (measured, not illustrative): the Flow Appearance Lab artifact
- `apps/flow/src/features/settings/FlowSettings.tsx:455` current section
- `apps/flow/src/layouts/FlowLayout.tsx:498` current accent writer
- `apps/flow/src/shared/lib/colorUtils.ts:86` `accentTints`
- `apps/flow/src/globals.css:103` hardcoded `--accent-fg`
- `packages/utils/src/hooks/useTheme.ts` unused `createThemeStore`
