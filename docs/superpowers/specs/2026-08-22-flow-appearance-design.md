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
| Panel scope       | All seven controls                            | Density and corners carry real cost (section 7). Flagged, reaffirmed, sequenced last.                                                                                             |
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

### 4.2 `appearanceTokens(prefs, resolvedMode)` (packages/utils)

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

**No preview pane.** The panel lives inside Flow, so the application is the
preview and every change applies live. The prototype needed one only because it
runs outside the app. Building one here would be duplicated surface with no
information the surrounding screen does not already carry.

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

## 7. Density and corners: cost, and sequencing

This is the largest single piece of the work and the only one with genuine
unknown cost. Flow's spacing and radii are literal Tailwind classes across
hundreds of components, so `--flow-u` and `--flow-r` have nothing to consume them
until those surfaces are migrated.

Scope the migration to the surfaces that constitute Flow's screens: list rows,
cards, chips, and nav items. That is a bounded set and it covers the great
majority of what a user sees.

**Sequence these two controls last**, so they can be dropped without unpicking
anything else if the estimate comes back worse than expected. A density control
that visibly moves only part of the app is the misleading-control problem: the
concept is real and has nowhere to fully land, so it would need to be disabled
with the reason stated rather than shipped half-working.

## 8. Testing

- **Contrast sweep.** Every palette times every accent fixture, asserting
  `--accent-fg` and `--flow-on-accent` both clear 4.5:1. The fixture list is the
  ten curated swatches **plus four deliberately dark hexes** (a navy, a plum, a
  forest, a brick), which is what makes it 8 times 14 rather than 8 times 10. The
  dark four are the only values that exercise the `--flow-on-accent` failure at
  all, so a sweep over curated swatches alone would pass while proving nothing.
  This is the test that makes the accessibility claim real.
- **Curated pairs as oracle.** The hand-checked foregrounds are the regression
  test for the derivation: point the algorithm at them and assert it reproduces
  them within tolerance.
- **Mutation probe on the derivation gate.** Force the predicate to a constant
  both ways and confirm the suite fails each time. An assertion over a dead
  branch passes for the wrong reason.
- **Store migration.** A stored `flow-theme` value survives the move to the new
  store.
- **Provider applies before paint.** No flash of the default palette on load.
- **`accent: null` falls back to the workspace colour**, and a stored accent
  overrides it.

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
