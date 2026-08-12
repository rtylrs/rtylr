# rtylr docs design QA

## Evidence

- Source visual truth: `qa-source-desktop.png` from `https://docs.retool.com/`.
- Combined comparison input: `qa-comparison-desktop.png` (source left, rtylr implementation right).
- Implementation desktop: `qa-implementation-desktop.png`.
- Implementation docs page: `qa-implementation-docs-desktop.png`.
- Implementation mobile: `qa-implementation-mobile.png`.
- Implementation mobile navigation: `qa-implementation-mobile-menu.png`.
- Desktop comparison viewport: 1022 × 719 CSS px; source and implementation screenshots are both 1022 × 719 pixels, 1× density.
- Mobile verification viewport: 390 × 844 CSS px; implementation screenshot is 390 × 844 pixels, 1× density.
- State: dark theme, home page for the full-view comparison; `/self-hosting/` for content-page verification.

## Findings

No actionable P0, P1, or P2 findings remain.

The implementation intentionally adapts the reference instead of copying it 1:1: rtylr branding and content replace Retool content, the rtylr grid illustration replaces the Retool artwork, the footer is rtylr-specific, and the reference hosting/version bottom bar is omitted per product direction.

## Focused checks

- Typography and hierarchy: dark display heading, compact uppercase navigation, muted body copy, and docs-page h1/h2 rhythm match the reference direction while using the workspace’s available font stack.
- Spacing and layout: desktop two-column hero, three-column docs layout, left navigation, right in-page navigation, mobile single-column layout, and full-height mobile drawer verified visually.
- Colors and tokens: implementation body background measured as `rgb(16, 16, 16)` (`#101010`); borders, muted text, blue CTA, and green announcement accent are stable across routes.
- Asset fidelity: existing rtylr logo is served from the Next root `public/` path and loaded with a non-zero natural width in the static build.
- Copy and content: self-hosting content is rtylr-specific, with `Self Host rtylr` present in both docs and landing footers.

## Primary interactions tested

- Desktop Build dropdown opens and exposes rtylr-specific destinations.
- Search opens, accepts `self host`, and returns the self-hosting overview.
- Footer `Self Host rtylr` navigates to `/self-hosting/`.
- Self-hosting page in-page navigation emits working heading anchors.
- Mobile menu opens as a full-height drawer at 390 × 844.
- Static build console/error check: no errors or warnings reported by the final browser tab.

## Comparison history

1. Missing heading IDs in Markdown output: added explicit `marked` heading IDs so the right-hand table of contents anchors work.
2. Broken logos after the VitePress-to-Next move: copied existing rtylr public assets to Next’s root `public/` path.
3. Mobile drawer constrained to the header: removed the header blur containing block so the drawer fills the viewport.
4. Retool-style bottom bar: removed intentionally after product direction; it is not part of the final rtylr UI.

## Implementation checklist

- [x] Desktop source/implementation comparison captured.
- [x] Mobile implementation and drawer captured.
- [x] Focused navigation, search, footer link, and in-page anchor checks completed.
- [x] Typecheck and static Next build passed.

final result: passed
