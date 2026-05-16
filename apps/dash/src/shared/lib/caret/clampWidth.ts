/**
 * Clamp the Caret panel width to a sensible window-relative range so the
 * resize handle can't push the user into a 50%-width sidebar.
 */
export function clampPanelWidth(px: number, windowWidth: number): number {
  const min = 320;
  const max = Math.min(800, Math.max(min, Math.floor(windowWidth * 0.6)));
  return Math.min(Math.max(Math.round(px), min), max);
}
