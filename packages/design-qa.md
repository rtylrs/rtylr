# Caret dark brutalist redesign QA

## Scope and visual direction

Caret is a persistent business-assistant sidecar, not a standalone destination. The final direction combines Dia's quiet focus-panel structure, Raycast's explicit history model, and VS Code/Cursor-style concurrent chat tabs with the user's requested modern-brutalist treatment: near-black surfaces, exposed one-pixel structure, sharp geometry, restrained teal status signals, and no decorative gradients or soft panel shadows.

Research sources:

- Dia panel and focus mode: `https://www.diabrowser.com/` and `https://www.diabrowser.com/changelog/1-5-0`
- Comet assistant sidecar: `https://www.perplexity.ai/help-center/comet/en/articles/11734688-assistant-panel`
- VS Code chat sessions: `https://code.visualstudio.com/docs/agents/run/sessions/manage-sessions`
- Cursor agent tabs and voice: `https://cursor.com/changelog/3-0` and `https://cursor.com/changelog/3-1`
- Raycast chat and history: `https://manual.raycast.com/ai/chat` and `https://www.raycast.com/blog/the-new-raycast`
- WCAG 2.2 contrast guidance: `https://www.w3.org/TR/WCAG22/`

## Evidence

Source visual truth:

- User-provided Dia reference: `/Users/mhomsi/.t3/userdata/attachments/a3db4e10-258b-439e-909c-85d054a03fde-3abaa1eb-6ea8-40cd-befd-20e2781dd3fe.png` (495 x 1067 px).
- Normalized Dia panel: `/tmp/caret-dark-research-2026-08-09/dia-panel-normalized.png` (372 x 800 px).
- Raycast AI chat reference: `/tmp/caret-dark-research-2026-08-09/raycast-ai-chat.png` (2152 x 1832 px).
- Normalized Raycast history reference: `/tmp/caret-dark-research-2026-08-09/raycast-history-normalized.png` (380 x 800 px).
- User-reported mode-picker defect: `/Users/mhomsi/.t3/userdata/attachments/a3db4e10-258b-439e-909c-85d054a03fde-ee38f040-84b5-41d7-80cc-30fea5320fda.png` (436 x 794 px).

Final implementation captures:

- Empty chat panel: `/tmp/caret-dark-research-2026-08-09/implementation-empty-panel-final.png` (380 x 800 px).
- Populated chat panel at rest: `/tmp/caret-dark-research-2026-08-09/implementation-chat-panel-final.png` (380 x 800 px).
- Populated chat with response hover: `/tmp/caret-dark-research-2026-08-09/implementation-chat-panel-hover.png` (380 x 800 px).
- Final live populated state: `/tmp/caret-dark-research-2026-08-09/implementation-chat-panel-final2.png` (380 x 800 px).
- History: `/tmp/caret-dark-research-2026-08-09/implementation-brutalist-history-380.png` (1280 x 800 px full viewport; 380 px panel).
- Settings: `/tmp/caret-dark-research-2026-08-09/implementation-brutalist-settings-380.png` (1280 x 800 px full viewport; 380 px panel).
- Compact mode picker: `/tmp/caret-dark-research-2026-08-09/implementation-mode-menu-final.png` (380 x 800 px).

Combined source-and-implementation comparison inputs:

- Dia empty panel vs. Caret empty panel: `/tmp/caret-dark-research-2026-08-09/compare-dia-caret-empty-final.png` (776 x 800 px).
- Raycast history vs. Caret history: `/tmp/caret-dark-research-2026-08-09/compare-raycast-caret-history.png` (784 x 800 px).
- Dia header vs. Caret header and tabs: `/tmp/caret-dark-research-2026-08-09/compare-header-tabs-focused.png` (760 x 461 px).
- Broken vs. corrected mode picker: `/tmp/caret-dark-research-2026-08-09/compare-mode-picker-before-after.png` (840 x 800 px).
- Copy action at rest vs. response hover: `/tmp/caret-dark-research-2026-08-09/compare-copy-rest-hover.png` (784 x 800 px).

## Normalization and state

- Browser viewport: 1280 x 800 CSS px.
- Primary Caret panel: 380 x 800 CSS px; responsive inspections also covered the 320 px minimum and 700 px expanded widths.
- T3 browser recordings were captured at 2x device scale, cropped to the right-side panel, then normalized to 380 x 800 px for comparison.
- The Dia source was proportionally normalized to 372 x 800 px. It and the final Caret capture use the same empty-sidecar state; Caret intentionally adds the requested concurrent-chat tab rail.
- History comparisons use the same search-and-select workflow. The Raycast source contains folders and pinning outside this task's scope; Caret preserves the simpler recent-chat archive.

## Required fidelity surfaces

- Typography: the inherited product font remains legible and compact. Chat body text is 13.5-14 px with a 1.5-1.6 line height; tab and metadata labels are subordinate without becoming low-contrast decoration.
- Spacing and layout: the 44 px header, compact 36-38 px tab rail, full-height content canvas, and persistent 50 px composer reproduce the reference sidecar hierarchy. Tabs remain content-sized instead of stretching into a second toolbar.
- Colors and tokens: the palette is limited to near-black surfaces, off-white text, neutral borders, and teal for focus, send, listening, selection checks, and active switches. There is no blue accent, gradient chrome, or soft panel/composer shadow.
- Geometry: panel and rail edges are square; tabs use a 2 px radius, message bubbles and controls use 4 px, and the composer uses 8 px to preserve input affordance without returning to pill-heavy styling.
- Copy and content: prompts and dynamic chat content remain unchanged. History and settings labels are concise. The empty state is deliberately silent rather than filling the canvas with promotional prompts.
- Icons and assets: existing Caret raster logo assets and Lucide icons are retained. No emoji, placeholder artwork, CSS-drawn icon, handcrafted SVG, or unrelated downloaded asset was added to the product.
- States: active tabs use a full one-pixel inset boundary rather than a bottom bar. User messages use a neutral bordered surface. History has a dedicated full view. The mode picker is a compact 168 px, one-line menu aligned to the trigger, with `menuitemradio` semantics and a checked selected state.
- Response actions: copy/retry controls are visually absent at rest, reveal only while the assistant response is hovered, and also reveal on keyboard focus. Hidden actions remain keyboard reachable and do not intercept pointer input.
- Responsiveness: tabs scroll only when required; follow-ups wrap; history titles truncate; composer controls remain intact; and the compact mode menu stays within the panel at narrow widths. No horizontal document overflow was observed.
- Accessibility: tabs expose `tablist`, `tab`, and `tabpanel`; arrow, Home, End, and Delete behavior is supported; settings rows expose accessible switches; menu selection exposes `aria-checked`; visible focus rings and reduced-motion handling remain. Escape closes the mode menu, restores focus to its trigger, and no longer closes the whole Caret panel.
- Contrast: measured ratios are 16.57:1 for primary text on the panel, 8.74:1 for muted text, 5.80:1 for dim text, 10.13:1 for teal on the panel, and 9.95:1 for dark text on teal. These exceed the WCAG 2.2 AA text thresholds used for this pass.

## Interaction and runtime checks

- Switched among three open chat tabs; verified independent tab content, tabpanel semantics, and restored scroll position.
- Verified ArrowLeft/ArrowRight/Home/End tab navigation and Delete-to-close behavior.
- Opened the dedicated history view, searched for `cheese`, verified filtering, cleared search, selected a previous chat, and returned to the active tab set.
- Entered and cleared composer text; verified send enable/disable behavior.
- Triggered dictation in the collaborative browser. Browser microphone permission was denied, and Caret displayed a clear dismissible permission error instead of duplicating interim speech text.
- Verified copy actions at rest, pointer hover, and keyboard focus. Resting opacity/pointer state was `0`/`none`; hover and focus resolved to `1`/`auto`.
- Opened the mode menu, selected another mode, restored Analyst, and verified the selected radio state. Escape closes only the menu and leaves the 380 px Caret panel open.
- Opened settings and toggled behavior rows; inspected history, settings, populated chat, and empty chat at 380 px, plus chat behavior at 320 px and 700 px.
- T3 browser snapshots reported no runtime diagnostics; no horizontal page overflow was present.

## Comparison history

### Pass 1 - blocked

- [P1] The first light/native pass contradicted the requested dark sidecar direction and made Caret feel visually absorbed by the dashboard.
- [P2] Soft radii, shadowed surfaces, and decorative accent treatments weakened the user's preferred brutalist character.
- [P2] History appeared as an oversized hover overlay, while concurrent chats looked like detached browser tabs.

Fixes: restored an independent dark surface; separated active tabs from archived history; removed light/blue fallbacks, bottom-tab bars, broad shadows, and promotional empty-state cards; and rebuilt the structure around exposed borders and compact controls.

### Pass 2 - blocked

- [P2] The initial dark correction still leaned toward soft macOS styling rather than the user's stated brutalist preference.
- [P2] Message actions remained permanently visible and read as a repeated action bar.

Fixes: reduced radii, flattened elevation, strengthened one-pixel geometry, made teal functional only, and changed response actions to hover/focus reveal.

### Pass 3 - blocked

- [P2] The composer mode picker opened as a 220 px settings card with multi-line descriptions, clipped against the panel edge, and obscured too much conversation content.
- [P2] Escape closed both the menu and the entire Caret panel.

Fixes: aligned the menu to the trigger's end edge, reduced it to compact one-line 168 px rows, added an explicit selected check/radio state, constrained it to the panel container, and consumed Escape before the global panel shortcut.

### Pass 4 - passed

The final combined comparisons show a quiet dark sidecar, explicit concurrent tabs without bottom bars, a dedicated history view, a compact composer and mode menu, and hover-only response actions. No actionable P0, P1, or P2 visual, interaction, responsive, or accessibility issue remains.

## Automated verification

- Caret tests: 9 files, 13 tests passed after the final mode-menu Escape test was added.
- Caret package typecheck: passed.
- Full monorepo typecheck: 20 tasks passed across 26 packages in scope.
- Full lint: 0 errors and 329 pre-existing warnings; no new Caret warning remains.
- Dashboard production build: passed; 4,925 modules transformed.
- `git diff --check`: passed for the Caret and QA-report scope; untracked additions were checked separately.

## Follow-up note

Live microphone transcription still depends on browser and operating-system permission. Permission-denied and repeated-interim-result behavior are covered, but hardware capture could not be completed in the collaborative browser.

final result: passed
