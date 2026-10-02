/**
 * Design tokens for the Caret panel. Kept here so themes can swap them at
 * the panel root via CSS variables.
 */

export const caretTokens = {
  accent: "var(--caret-accent, #63d1c4)",
  accentStrong: "var(--caret-accent-strong, #7cddd2)",
  accentSoft: "var(--caret-accent-soft, rgba(100, 197, 185, 0.12))",
  accentBorder: "var(--caret-accent-border, rgba(100, 197, 185, 0.28))",
  accentInk: "var(--caret-on-accent, #0d1715)",
  bg: "var(--caret-bg, #121315)",
  surface: "var(--caret-surface, #17181b)",
  surfaceHigh: "var(--caret-surface-high, #1e2023)",
  surfaceRaised: "var(--caret-surface-raised, #18191c)",
  border: "var(--caret-border, rgba(255, 255, 255, 0.12))",
  borderStrong: "var(--caret-border-strong, rgba(255, 255, 255, 0.24))",
  text: "var(--caret-text, #f2f2ef)",
  textMuted: "var(--caret-text-muted, #b2b2ae)",
  textDim: "var(--caret-text-dim, #90908c)",
  inputBg: "var(--caret-input-bg, #18191c)",
  hover: "var(--caret-hover, rgba(255, 255, 255, 0.055))",
  active: "var(--caret-active, rgba(255, 255, 255, 0.09))",
  shadow: "var(--caret-shadow, 4px 4px 0 rgba(0, 0, 0, 0.44))",
  danger: "var(--caret-danger, #ff7b82)",
  dangerSoft: "var(--caret-danger-soft, rgba(255, 123, 130, 0.12))",
  dangerText: "var(--caret-danger-text, #ff9ca1)",
  warn: "var(--caret-warn, #e7b569)",
  warnSoft: "var(--caret-warn-soft, rgba(231, 181, 105, 0.12))",
  success: "var(--caret-success, #63c69d)",
  thinking: "var(--caret-thinking, #64c5b9)",
};

export const caretRadius = {
  sm: "var(--caret-radius-sm, 2px)",
  md: "var(--caret-radius-md, 4px)",
  lg: "var(--caret-radius-lg, 6px)",
  xl: "var(--caret-radius-xl, 8px)",
};
