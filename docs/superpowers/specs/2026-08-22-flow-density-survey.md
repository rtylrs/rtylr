# Flow density, text size, and corners: cost survey

Date: 2026-08-22
Status: recon, read-only. No component was modified.
Scope surveyed: `apps/flow/src` (312 `.ts`/`.tsx` files, 111 of them `.tsx`)
Answers: section 7 of `docs/superpowers/specs/2026-08-22-flow-appearance-design.md`

## 0. Headline

Section 7 is wrong, but it is wrong in the direction nobody expected. It assumes
literal Tailwind classes cannot respond to a custom property, so `--flow-u` and
`--flow-r` need hundreds of call sites rewritten. **In Tailwind v4 as this repo
configures it, they already do respond.** `p-3` does not compile to
`padding: 12px`. It compiles to `padding: calc(var(--spacing) * 3)`. `rounded-lg`
compiles to `border-radius: var(--radius-lg)`. Both tokens are ordinary custom
properties on `:root`, and Flow's `globals.css` already overrides tokens of
exactly that kind at `globals.css:177`.

| Control       | Verdict                       | The one number                                                                                             |
| ------------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------- |
| **Density**   | Ship. Roughly one line of CSS | **1368 of 1409** spacing declarations (97.1%) already resolve through `var(--spacing)`                     |
| **Corners**   | Ship. Half a day              | **170 of 335** radius declarations already resolve through `var(--radius-*)`; the other 165 are one rename |
| **Text size** | Cut, or show disabled         | **618 of 637** font-size declarations (97.0%) are px literals that no token can reach                      |

The migration section 7 describes is not needed for two of the three controls,
and does not solve the third. The bounded four-category scope it proposes is also
the wrong shape: measured against what a user sees on the main screens, it would
move **16% to 27%** of the surface, which is well under half and exactly the
misleading-control failure the spec was trying to avoid.

## 1. Method, and how to reproduce it

Every count below comes from ripgrep 15.2.0 with `--pcre2`, run from
`/Users/abedamouneh/Development/rtylr/apps/flow/src`. Patterns are single-quoted
so zsh cannot eat them, and every family pattern is bounded by a negative
lookbehind `(?<![\w-])` and a negative lookahead `(?![\w.-])` so that identifiers
and unrelated hyphenated strings do not inflate the totals.

That guard is not decorative. The first pass without it counted `my-tasks` (28),
`my-issues` (17), `p-web` (2) and `roundedPreview` (4) as spacing and radius
utilities. Every number in this document is from the guarded pass.

The spacing family pattern, referred to below as `$SP`:

```
(?<![\w-])(?:p|px|py|pt|pr|pb|pl|ps|pe|m|mx|my|mt|mr|mb|ml|ms|me|gap|gap-x|gap-y|space-x|space-y)-(?:[0-9]+(?:\.5)?|px|\[[^\]]+\])(?![\w.-])
```

Two counting modes are used and they are not interchangeable. `rg -o … | wc -l`
counts **occurrences**. `rg -c` counts **lines containing at least one match**,
which is what the hotspot tables in sections 1.4 report. Note also that ripgrep's
`-I` is `--no-filename`, not `--no-ignore`: an early `rg -cI` produced a per-file
table with the filenames stripped, which is why the tables were regenerated
without it.

### 1.1 Spacing

```bash
rg -o  --pcre2 --no-filename "$SP" -g '*.tsx' -g '*.ts' -g '*.jsx' . | sort | uniq -c | sort -rn
rg -l  --pcre2 "$SP" -g '*.tsx' -g '*.ts' -g '*.jsx' . | wc -l
```

**1409 occurrences, 116 distinct utility-plus-value pairs, across 74 files.**

By family:

| Family                                                                                   | Occurrences |
| ---------------------------------------------------------------------------------------- | ----------: |
| `gap-*`                                                                                  |         317 |
| `px-*`                                                                                   |         294 |
| `py-*`                                                                                   |         290 |
| `mb-*`                                                                                   |         151 |
| `p-*`                                                                                    |         141 |
| `mt-*`                                                                                   |          91 |
| `space-y-*`                                                                              |          70 |
| all others (`pt`, `pb`, `ml`, `m`, `mx`, `ms`, `pl`, `pr`, `me`, `mr`, `gap-x`, `gap-y`) |          55 |

The distribution by **value** is the number section 7 needed, and it is
concentrated far more tightly than "hundreds of components" implies:

| Scale step                  | Occurrences | Share | Cumulative |
| --------------------------- | ----------: | ----: | ---------: |
| `2`                         |         325 | 23.1% |      23.1% |
| `3`                         |         254 | 18.0% |      41.1% |
| `1.5`                       |         194 | 13.8% |      54.9% |
| `1`                         |         193 | 13.7% |      68.6% |
| `4`                         |         136 |  9.7% |      78.2% |
| `6`                         |         101 |  7.2% |      85.4% |
| `0.5`                       |          65 |  4.6% |      90.0% |
| `2.5`                       |          40 |  2.8% |      92.8% |
| `5`                         |          28 |  2.0% |      94.8% |
| `px`                        |          16 |  1.1% |      96.0% |
| `8`                         |          13 |  0.9% |      96.9% |
| `0`                         |          13 |  0.9% |      97.8% |
| `9`, `3.5`, `7`, `20`, `10` |          19 |  1.4% |      99.1% |
| 6 arbitrary values          |          12 |  0.9% |     100.0% |

Six values (`1`, `1.5`, `2`, `3`, `4`, `6`) account for 85.4% of every spacing
declaration in the app. Eight cover 92.8%. This is the "handful of values
dominate" condition the task asked about, and it is met comfortably.

The twelve arbitrary values are `gap-[18px]` (5), `mb-[7px]` (2), `gap-[26px]`
(2), `py-[15px]` (1), `pt-[10vh]` (1), and
`pb-[max(0.5rem,env(safe-area-inset-bottom))]` (1). The last one is a safe-area
inset and must stay literal.

There is exactly one negative margin in the app (`-ml-1`), and 70 occurrences of
responsive-prefixed spacing (`sm:`, `md:`, `lg:`). Both behave identically to the
unprefixed forms under a token swap, so neither is a complication.

### 1.2 Radii

```bash
rg -o --pcre2 --no-filename '(?<![\w-])rounded(?:-(?:t|r|b|l|s|e|tl|tr|br|bl|ss|se|es|ee))?(?:-(?:none|sm|md|lg|xl|2xl|3xl|full|\[[^\]]+\]|[0-9]+))?(?![\w.-])' -g '*.tsx' -g '*.ts' -g '*.jsx' .
```

**418 occurrences, 9 distinct forms, across 63 files.** The concentration is
extreme: four forms cover 99.3%.

| Form            | Occurrences | Compiles to                 | Responds to a runtime token? |
| --------------- | ----------: | --------------------------- | ---------------------------- |
| `rounded`       |         162 | `border-radius: 0.25rem`    | **No** (see section 2.1)     |
| `rounded-full`  |          83 | `calc(infinity * 1px)`      | No, and should not           |
| `rounded-lg`    |          77 | `var(--radius-lg)`          | **Yes**                      |
| `rounded-md`    |          57 | `var(--radius-md)`          | **Yes**                      |
| `rounded-xl`    |          34 | `var(--radius-xl)`          | **Yes**                      |
| `rounded-[9px]` |           2 | `9px`                       | No                           |
| `rounded-t-xl`  |           1 | `var(--radius-xl)` (top)    | **Yes**                      |
| `rounded-b-xl`  |           1 | `var(--radius-xl)` (bottom) | **Yes**                      |
| `rounded-t`     |           1 | `0.25rem` (top)             | No                           |

`rounded-sm` and `rounded-2xl` have **zero** occurrences in `apps/flow/src`. That
absence turns out to matter (section 4.2).

Excluding `rounded-full`, which paints pills on avatars and chips and must stay a
pill at every corner setting, the addressable set is 335 declarations. **170 of
them (50.7%) already read a token.** The remaining 165 are `rounded` (162),
`rounded-t` (1) and `rounded-[9px]` (2).

Flow's own `globals.css` contains exactly one radius declaration,
`border-radius: 9999px` at line 231, on a scrollbar thumb. It is a pill and is
out of scope.

### 1.3 Type

```bash
rg -o --pcre2 --no-filename '(?<![\w-])text-\[[0-9]+(?:\.[0-9]+)?px\]' -g '*.tsx' -g '*.ts' .   # 618
rg -o --pcre2 --no-filename '(?<![\w-])text-(?:xs|sm|base|lg|xl|2xl|3xl)(?![\w.-])' -g '*.tsx' -g '*.ts' .  # 19
rg -l --pcre2 '(?<![\w-])text-\[[0-9]' -g '*.tsx' -g '*.ts' . | wc -l   # 61
```

**618 px-literal font sizes across 61 files, against 19 uses of the named scale.**
Px literals are 97.0% of Flow's type. The distinct sizes:

| Size         | Occurrences |   At x0.92 | Under the 10px floor?                      |
| ------------ | ----------: | ---------: | ------------------------------------------ |
| `11px`       |         189 |    10.12px | No, by 0.12px                              |
| `12px`       |         154 |    11.04px | No                                         |
| **`10px`**   |     **103** | **9.20px** | **Yes**                                    |
| `13px`       |          96 |    11.96px | No                                         |
| `15px`       |          22 |    13.80px | No                                         |
| `16px`       |          18 |    14.72px | No                                         |
| **`9px`**    |      **10** | **8.28px** | **Yes, and already below the floor today** |
| `14px`       |          10 |    12.88px | No                                         |
| `20px`       |           4 |    18.40px | No                                         |
| `22px`       |           3 |    20.24px | No                                         |
| `18px`       |           3 |    16.56px | No                                         |
| `12.5px`     |           2 |    11.50px | No                                         |
| **`9.5px`**  |       **1** | **8.74px** | **Yes, and already below the floor today** |
| `24px`       |           1 |    22.08px | No                                         |
| `17px`       |           1 |    15.64px | No                                         |
| **`10.5px`** |       **1** | **9.66px** | **Yes**                                    |

`globals.css:98` states the floor: "WCAG's large-text allowance never applies in
Flow: the type floor is 10px."

**115 occurrences breach that floor at a 0.92 multiplier.** 104 of them are legal
today and would become illegal. The other 11 already breach it at multiplier 1.0:
ten `text-[9px]` and one `text-[9.5px]`, at
`layouts/FlowLayout.tsx:717`, `layouts/FlowLayout.tsx:805`,
`shared/components/CommandBar.tsx:326`, `shared/components/CommentInput.tsx:116`,
`features/calendar/Calendar.tsx:212`, `features/calendar/Calendar.tsx:240`,
`features/projects/ProjectOverview.tsx:248`, `features/projects/ProjectList.tsx:142`,
`features/insights/Insights.tsx:155`,
`features/workspace-dashboard/WorkspaceDashboard.tsx:470`, and
`features/time/AddHours.tsx:621`. That is a pre-existing defect this survey found
incidentally, not a consequence of the appearance work. Most of those are single
initials inside a 20px avatar circle, so the practical severity is low, but the
count is real and it means the floor is aspirational rather than enforced.

There are also **10 inline `fontSize` values in JS**, which no CSS token or root
font-size can reach: `shared/components/CycleBadge.tsx:27` (10 or 11),
`shared/components/TaskRow.tsx:142` (10),
`shared/components/LabelChip.tsx:19` (10 or 11),
`shared/components/AssigneeAvatar.tsx:27` and `:66` (9 and 10),
`shared/components/TaskContextMenu.tsx:57` (12),
`router.tsx:64`, `routes/AppRoutes.tsx:46`, `main.tsx:31`,
`shared/components/PreviewRouteUnavailable.tsx:24`. Three of those six components
are chips and rows, which is to say the four-category migration set has a
JS-inline type problem inside it.

Flow uses 12 inline `padding`/`margin`/`gap`/`borderRadius` values in JS as well
(8 `padding`, 6 `margin`, 3 `borderRadius`, 2 `paddingTop`, 1 `paddingLeft`,
1 `gap`). Small, but they are immune to any token approach and must be handled by
hand if density and corners are to be complete.

### 1.4 Where the work is concentrated

Lines containing at least one match, `rg -c`, top files:

| File                                                  | Spacing lines | Type lines | Radius lines |
| ----------------------------------------------------- | ------------: | ---------: | -----------: |
| `features/settings/FlowSettings.tsx`                  |            72 |         57 |           40 |
| `layouts/FlowLayout.tsx`                              |            56 |         28 |           30 |
| `features/projects/ProjectSettings.tsx`               |            48 |         37 |           23 |
| `features/tasks/TaskDetail.tsx`                       |            42 |         34 |           17 |
| `features/time/AddHours.tsx`                          |            39 |         28 |           20 |
| `features/projects/components/BoardSettingsPanel.tsx` |            35 |         18 |           14 |
| `features/tasks/components/TimePanel.tsx`             |            28 |         23 |           15 |
| `features/settings/sections/TimeSection.tsx`          |            28 |         17 |            0 |
| `features/workspace-dashboard/WorkspaceDashboard.tsx` |            27 |         21 |           17 |

33 of the 74 spacing-bearing files hold 80% of the occurrences. If a per-call-site
migration were ever needed, it would be a 33-file job for the bulk and a 74-file
job for completeness.

## 2. The Tailwind v4 constraint, which turns out to be a Tailwind v4 gift

This is the make-or-break question, so it was answered by compiling rather than
by reading. `apps/flow/package.json:43` pins `tailwindcss@^4.2.4` and
`node_modules/tailwindcss/package.json` reports 4.2.4 installed.

### 2.1 What the compiler actually emits

A probe file was compiled with the real plugin and the real shared theme, using
the repo's own `postcss` and `@tailwindcss/postcss` resolved through
`createRequire` against the repo root. Input:

```css
@import '.../node_modules/tailwindcss/index.css';
@import '.../packages/tailwind-config/theme.css';
@source "./probe.html";
```

with a probe element carrying
`p-3 px-3 gap-2 py-1.5 rounded rounded-lg rounded-md rounded-xl rounded-full text-[12px] text-xs p-[var(--flow-u)] px-[calc(var(--flow-u)*3)] rounded-[var(--flow-r)] text-[length:calc(12px*var(--flow-fs))] w-4 h-4 w-64 min-h-9 top-2 translate-x-2 inset-2 -mt-2 leading-4`.

The emitted `@layer utilities`, verbatim in the parts that matter:

```css
.p-3            { padding: calc(var(--spacing) * 3); }
.px-3           { padding-inline: calc(var(--spacing) * 3); }
.py-1\.5        { padding-block: calc(var(--spacing) * 1.5); }
.gap-2          { gap: calc(var(--spacing) * 2); }
.-mt-2          { margin-top: calc(var(--spacing) * -2); }
.w-4            { width: calc(var(--spacing) * 4); }
.h-4            { height: calc(var(--spacing) * 4); }
.w-64           { width: calc(var(--spacing) * 64); }
.min-h-9        { min-height: calc(var(--spacing) * 9); }
.top-2          { top: calc(var(--spacing) * 2); }
.inset-2        { inset: calc(var(--spacing) * 2); }
.translate-x-2  { --tw-translate-x: calc(var(--spacing) * 2); … }
.leading-4      { --tw-leading: calc(var(--spacing) * 4); line-height: calc(var(--spacing) * 4); }
.rounded-lg     { border-radius: var(--radius-lg); }
.rounded-md     { border-radius: var(--radius-md); }
.rounded-xl     { border-radius: var(--radius-xl); }
.rounded-full   { border-radius: calc(infinity * 1px); }
.rounded        { border-radius: 0.25rem; }
.text-xs        { font-size: var(--text-xs); line-height: var(--tw-leading, var(--text-xs--line-height)); }
.text-\[12px\]  { font-size: 12px; }
```

And the emitted `@layer theme` re-declares the tokens on `:root, :host`:

```css
:root,
:host {
  --spacing: 0.25rem;
  --text-xs: 0.75rem;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;
}
```

So the answer to "can a custom property be threaded into v4 utilities in this
setup" is: **the threading already exists, and it is the default.** No arbitrary
value syntax is required. `p-[var(--flow-u)]` does work (it compiles to
`padding: var(--flow-u)`), but it is the expensive path and it is unnecessary.

Two mechanics worth recording because they cost time to establish:

- `--spacing` and `--radius-*` live in plain `@theme` blocks, so v4 emits them as
  real custom properties on `:root` and utilities reference them with `var()`.
- Bare `rounded` comes from v4's `@theme default inline reference` block
  (`node_modules/tailwindcss/theme.css:503-510`, `--radius: 0.25rem`). `inline`
  means the value is baked into the rule at build time. That is why `rounded` is
  the one radius form that cannot be moved at runtime.
- The unquoted-custom-property form `p-[--flow-u]` is a trap: it compiles to the
  literal `padding: --flow-u`, which is invalid. Always write
  `p-[var(--flow-u)]`.

### 2.2 The cascade, verified in a browser rather than argued from the spec

The token swap only works if an unlayered `:root` rule in `globals.css` beats the
`:root` inside `@layer theme`. A page was served on `127.0.0.1:8791` reproducing
the exact layer structure and read back with `getComputedStyle`:

```css
@layer theme, base, components, utilities;
@layer theme {
  :root,
  :host {
    --spacing: 0.25rem;
    --radius-lg: 12px;
    --text-xs: 0.75rem;
  }
}
@layer utilities {
  .p-3 {
    padding: calc(var(--spacing) * 3);
  }
  .rounded-lg {
    border-radius: var(--radius-lg);
  }
  .rounded {
    border-radius: 0.25rem;
  }
  .text-xs {
    font-size: var(--text-xs);
  }
}
:root {
  --spacing: 0.3125rem;
  --radius-lg: 2px;
  --text-xs: 0.875rem;
} /* unlayered */
```

Computed result:

```json
{ "padding": "15px", "radiusLg": "2px", "fontSize": "14px", "bareRounded": "4px" }
```

`p-3` moved from 12px to 15px. `rounded-lg` moved from 12px to 2px. `text-xs`
moved from 12px to 14px. Bare `rounded` did not move. That is the whole finding in
four values.

This is not a novel technique in this repo. `apps/flow/src/globals.css:177` already
does it, for colour:

```css
:root {
  --color-accent: var(--accent);
  --color-accent-foreground: var(--flow-on-accent);
  --color-ring: var(--accent);
  --color-muted: var(--flow-text-muted);
  /* … 10 more --color-flow-* */
}
```

That block is the existing, working, documented precedent for the exact mechanism
density and corners need. `packages/tailwind-config/theme.css:19-24` documents the
pattern in its own header comment. Nothing new has to be invented.

For completeness: a repo-wide search for the arbitrary-value form found only four
instances, all colour (`text-[var(--flow-text)]` twice in Flow,
`text-[var(--dash-text)]` and `text-[var(--dash-text-muted)]` in dash). So there
is **no** precedent for `p-[var(--flow-u)]`, and there does not need to be.

### 2.3 The one real hazard in the token swap

`--spacing` is a single token behind far more than padding and gap. In
`apps/flow/src` it also drives **447** further declarations:

| Family                                                                                                             | Occurrences |
| ------------------------------------------------------------------------------------------------------------------ | ----------: |
| `h-*`                                                                                                              |         157 |
| `w-*`                                                                                                              |         136 |
| `min-h-*`                                                                                                          |          58 |
| `min-w-*`                                                                                                          |          44 |
| `start-*`, `top-*`, `inset-*`, `left-*`, `translate-x-*`, `max-h-*`, `inset-x/y-*`, `end-*`, `right-*`, `bottom-*` |          52 |

So a `--spacing` change moves **1856** declarations, not 1409. That is mostly what
you want: it is why an icon box, an avatar and a row height all scale together
instead of a row's padding shrinking around a fixed-height icon. But it also
scales fixed layout widths. The value distribution says how much of a problem that
is: only 19 of the 395 `w`/`h`/`min-w`/`min-h` occurrences use a step of 14 or
above (`w-64` twice, `w-48`/`h-48` three times, `w-40` three times, and a handful
of 32, 24, 20, 16, 14). At compact (x0.85) `w-64` would go from 256px to 218px.

Two consequences follow, and neither is expensive:

1. Those 19 large layout dimensions should be pinned to literal `w-[256px]`-style
   values, or to a separate `--flow-shell-w` token, before density ships. That is
   a one-sitting change to a named, enumerable list: `w-48` (3), `w-24` (3),
   `h-40` (3), `w-64` (2), `w-32` (2), `w-16` (2), `h-14` (2), `w-60` (1),
   `w-20` (1).
2. Compact must not push a touch target under 44px. **53 fixed heights already sit
   below 44px at multiplier 1.0**: `h-7` (23 occurrences, 28px), `h-8` (21, 32px),
   `h-10` (5, 40px), `h-9` (4, 36px). Flow has no `min-h` floor to catch them:
   every one of its 58 `min-h-*` occurrences is `min-h-0`. Whether each of those
   53 is an interactive control was not audited, so treat the number as an upper
   bound on the problem rather than a defect count. Either way the floor is not
   currently enforced and density does not create the problem, but compact at
   x0.85 would take `h-8` to 27px and make an existing defect meaningfully worse.
   Audit those heights against the 44px rule as part of the density work, not
   after it.

## 3. The four core surfaces, and why the proposed scope does not work

### 3.1 What the four categories actually are

| Category  | Component                                                                                          | Spacing occ. | Radius occ. | Type occ. |
| --------- | -------------------------------------------------------------------------------------------------- | -----------: | ----------: | --------: |
| List rows | `apps/flow/src/shared/components/TaskRow.tsx`                                                      |            9 |           5 |         6 |
| Cards     | `apps/flow/src/shared/components/TaskCard.tsx`                                                     |           16 |           2 |         8 |
| Chips     | `apps/flow/src/shared/components/LabelChip.tsx`                                                    |            2 |           2 |         0 |
| Chips     | `apps/flow/src/shared/components/CycleBadge.tsx`                                                   |            1 |           1 |         0 |
| Chips     | `apps/flow/src/shared/components/StatusIcon.tsx`                                                   |            1 |           0 |         1 |
| Chips     | `apps/flow/src/shared/components/PriorityIcon.tsx`                                                 |            1 |           0 |         1 |
| Chips     | `apps/flow/src/shared/components/AssigneeAvatar.tsx`                                               |            0 |           2 |         0 |
| Nav items | `apps/flow/src/layouts/FlowLayout.tsx` lines 800-1060 (four `<nav>` blocks at 815, 886, 951, 1014) |           20 |           0 |         0 |

**Total: 50 spacing declarations. That is 3.5% of the app's 1409.**

### 3.2 The categories are not where the screens live

This is the part that sinks the proposal. The shared row and card components are
not used by the screens the spec names.

```bash
for c in TaskRow TaskCard LabelChip CycleBadge AssigneeAvatar StatusIcon PriorityIcon; do
  rg -l "components/$c'" -g '*.tsx' .
done
```

- `TaskRow` is imported by **`features/views/Views.tsx`** and
  **`features/triage/Triage.tsx`** only.
- `TaskCard` is imported by **`features/tasks/MyIssues.tsx`** and
  **`features/projects/ProjectBoard.tsx`**.
- `LabelChip` is imported by `features/tasks/TaskDetail.tsx` only.
- `features/inbox/Inbox.tsx` imports **none** of the seven. Its 24 spacing
  declarations and 12 px type sizes are inline markup.
- `features/focus-today/FocusToday.tsx` imports only `PriorityIcon`. Its 14
  spacing declarations are inline markup.

So "list rows" is not one component with two hundred call sites. It is one
component used on two secondary screens, plus hand-rolled row markup on Inbox and
Today.

### 3.3 What percentage of the main screens would move

Counting spacing declarations in each screen's full render tree, with the shell
(`FlowLayout.tsx`, 99 occurrences, of which 20 are the nav blocks):

| Screen                                               | Screen file | Shared components in tree | Shell | Total | Would move |   Share |
| ---------------------------------------------------- | ----------: | ------------------------: | ----: | ----: | ---------: | ------: |
| My Tasks (`features/tasks/MyIssues.tsx`)             |          17 |           16 (`TaskCard`) |    99 |   132 |         36 | **27%** |
| Project board (`features/projects/ProjectBoard.tsx`) |          37 |           16 (`TaskCard`) |    99 |   152 |         36 | **24%** |
| Today (`features/focus-today/FocusToday.tsx`)        |          14 |        1 (`PriorityIcon`) |    99 |   114 |         21 | **18%** |
| Inbox (`features/inbox/Inbox.tsx`)                   |          24 |                         0 |    99 |   123 |         20 | **16%** |

**Well under half, on every one of the four screens the spec names.** Between 16%
and 27%.

Two honest caveats on that reasoning. Declaration count is a proxy for visual
area, not visual area itself: a dense card row spends many declarations on few
pixels, and the shell spends few declarations on the rail width and header height.
That cuts against the migration rather than for it, though, because the shell's
handful of declarations are precisely the ones a density control is most visible
in: the rail width, the header height, the content gutter. A density control that
moves the card interiors and leaves the rail, the header and the gutters exactly
where they were would read as broken. Second, the shell is shared across all four
screens, so its 99 declarations are counted once per screen; that is correct for a
"what does the user see" question and would be wrong for a "how much work" one.

The conclusion stands either way: the bounded scope in section 7 is the
misleading-control problem, not the escape from it. Which is moot, because
section 2 says no migration is needed.

## 4. Verdicts

### 4.1 Density: ship it

The work is not a migration. It is:

1. `:root { --spacing: calc(0.25rem * var(--flow-density, 1)); }` in
   `apps/flow/src/globals.css`, in the same unlayered block as the existing
   `--color-*` re-point at line 177, with the multiplier written by
   `AppearanceProvider`. 1368 of 1409 spacing declarations move (97.1%). The 41
   that do not are 16 `px` hairlines, 13 zeroes, and 12 arbitrary values, and all
   three groups are correct to leave fixed.
2. Pin the 19 large `w-*`/`h-*` layout dimensions (step 14 and above) so the rail
   and panel widths do not scale with row density.
3. Audit the 53 sub-44px fixed heights (`h-7`, `h-8`, `h-9`, `h-10`) against the
   44px floor. They already fail it at multiplier 1.0, so this is a fix the app
   needs regardless; compact just makes ignoring it untenable.
4. Convert the 12 inline JS padding/margin/gap values, or accept them as fixed.

Estimate: one sitting for the token plus the pinning, and a second for the touch
target audit. That is a fraction of what section 7 budgeted, and item 3 is work
worth doing on its own merits.

### 4.2 Corners: ship it, and it is cheaper than density

170 of the 335 addressable radius declarations already read `var(--radius-md)`,
`var(--radius-lg)` or `var(--radius-xl)`. Re-point those three tokens in Flow's
`:root` and half the app's corners follow a control with zero call-site edits.
`rounded-full` (83) stays a pill, which is what you want on avatars and chips.

The 162 bare `rounded` need one decision. Two options were tested:

- **Rename `rounded` to `rounded-sm` (recommended).** `rounded-sm` has **zero**
  existing occurrences in `apps/flow/src`, so the rename cannot collide with
  anything. Set `--radius-sm: 4px` in Flow's `:root` to preserve today's exact
  appearance (v4's bare `rounded` is 0.25rem = 4px; the shared preset's
  `--radius-sm` is 6px, so without that line the rename would silently add 2px in
  162 places). One `sed` across 63 files, reviewable as a mechanical diff.
- **Override `.rounded` from an unlayered rule.** Verified working: an unlayered
  `.rounded { border-radius: var(--flow-r); }` computed to 2px against the layered
  built-in's 4px. But the same probe showed it also beats `.rounded-lg`, so
  `class="rounded rounded-lg"` would collapse to the bare value. No line in
  `apps/flow/src` currently carries both (checked: zero matches), but a multi-line
  `cn()` call could, and a future one certainly could. An `@utility rounded`
  override in `globals.css` does **not** work at all: v4 emits the custom rule
  before the built-in, so the built-in wins on source order.

Plus `rounded-t` (1) and `rounded-[9px]` (2) by hand.

Estimate: half a day, most of it reviewing a mechanical rename.

### 4.3 Text size: cut it, or ship it disabled with the reason stated

97.0% of Flow's type is px literals that no token can reach. `text-xs` and friends
would follow a `--text-*` swap, but there are 19 of them against 618 literals.

The three available approaches and why each fails:

| Approach                                      | Reaches the 618?                                                                   | Verdict                                                                                                                                                                  |
| --------------------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Root font-size plus rem                       | No. `text-[11px]` is immune, and the 19 rem-based uses are the only ones that move | Moves 3% of the type. Worse than nothing                                                                                                                                 |
| Rewrite to `text-[calc(11px*var(--flow-fs))]` | Yes                                                                                | 618 edits across 61 files, plus 10 inline JS `fontSize` values in 6 components. This is the real cost, and it is roughly ten times the density and corners work combined |
| `zoom` on the app root                        | Yes                                                                                | But it scales spacing, radii, borders and images too. That is a UI scale control, not a text size control, and it would collide with density                             |

And even if the rewrite were done, the floor bites: **115 of the 618 declarations
go under Flow's own 10px legibility floor at a 0.92 multiplier**, 103 of them from
`text-[10px]` alone. A "small" text setting that takes 103 labels to 9.2px in an
app whose `globals.css:98` says the floor is 10px is not a setting, it is a
regression with a switch on it. The asymmetric option (a "large" step only, no
"small") is coherent but is a different control from the one the spec designed,
and it still needs the 618-site rewrite to work at all.

Recommendation: **show Text size disabled, with the reason beside it**, using
`GateNotice` per the root CLAUDE.md rule on a control that cannot store its value.
The concept is real, it is standard in comparable products, and it has nowhere to
land, which is exactly the case that rule covers. Do not ship it working on 3% of
the type.

If it is later wanted for real, the honest sequencing is: first collapse the 16
distinct px sizes onto a named scale (`--text-*` entries in Flow's `@theme`, or
five semantic tokens), which is a worthwhile cleanup in its own right and fixes
the 11 sub-floor sizes on the way. The multiplier then becomes free, exactly as it
is for density and corners. That is the migration section 7 should have been
describing, and it is about type, not about spacing.

## 5. Where section 7 is wrong, plainly

- "Flow's spacing and radii are literal Tailwind classes … so `--flow-u` and
  `--flow-r` have nothing to consume them" is **false for this setup**. 97.1% of
  spacing and 50.7% of addressable radii already resolve through custom
  properties, because that is how Tailwind v4 compiles them and because the repo
  uses plain `@theme` blocks rather than `inline` ones. The compiled output in
  section 2.1 and the computed styles in section 2.2 are the evidence.
- "This is the largest single piece of the work and the only one with genuine
  unknown cost" is **inverted**. Density and corners are the cheapest two of the
  seven controls. Text size is the expensive one, and section 7 does not mention
  it at all.
- "Scope the migration to list rows, cards, chips, and nav items … it covers the
  great majority of what a user sees" is **wrong on its own terms**: those
  components are 3.5% of the app's spacing declarations and 16% to 27% of the four
  main screens. `Inbox.tsx` and `FocusToday.tsx` import none of them.
- "Sequence these two controls last, so they can be dropped" should be inverted.
  Density and corners are the two that should ship, and text size is the one to
  sequence out. Corners in particular should go early: it is the single cheapest
  visible control in the panel and it exercises the provider plumbing on 170
  declarations without touching a component.

## 6. What was counted and what was inferred

Counted, reproducibly, with the commands shown: every occurrence figure, file
count, distinct-value list, and per-file table in sections 1 and 3.

Verified by execution: the compiled Tailwind output in section 2.1 (real plugin,
real shared theme, tailwindcss 4.2.4) and the computed-style results in section
2.2 and 4.2 (served over HTTP, read via `getComputedStyle` in a browser).

Inferred, and flagged as such: the visual-share percentages in section 3.3 use
declaration count as a proxy for screen area, which is an approximation and is
argued about in that section. The effort estimates in section 4 are judgement, not
measurement. The claim that `--spacing` scaling row heights and icon boxes
together is desirable is a design opinion, not a finding.

Not attempted: running Flow in a browser to watch a live `--spacing` change move
the real UI. That needs the auth app on :5176 alongside Flow, and three sibling
agents are editing `packages/` concurrently, so a `pnpm install` and dev-server
boot would have raced them. The cascade probe in section 2.2 isolates the exact
mechanism instead, and `globals.css:177` is the same mechanism already working in
production. A live confirmation is a ten-minute task once `packages/` is quiet and
should be the first step of the implementation, not a reason to doubt the numbers.
