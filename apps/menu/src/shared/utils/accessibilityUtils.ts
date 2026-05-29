/**
 * Accessibility utility helpers for the Menu app.
 */

/** Return an aria-label string for a cart item count badge. */
export function cartCountAriaLabel(count: number): string {
  if (count === 0) return 'Cart is empty';
  return count === 1 ? '1 item in cart' : `${count} items in cart`;
}

/** Return the aria-pressed value for a toggle button. */
export function ariaPressed(isActive: boolean): 'true' | 'false' {
  return isActive ? 'true' : 'false';
}

/** Return the aria-expanded value for an expandable element. */
export function ariaExpanded(isOpen: boolean): 'true' | 'false' {
  return isOpen ? 'true' : 'false';
}

/** Return the role attribute for a given interactive context. */
export function buttonRole(isLink: boolean): 'button' | 'link' {
  return isLink ? 'link' : 'button';
}

/** Generate a stable DOM id from a label (lowercased, spaces to hyphens). */
export function labelToId(label: string): string {
  return label
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '');
}

/** Return true if the contrast ratio between two hex colors meets WCAG AA (4.5:1). */
export function isContrastSufficient(hexA: string, hexB: string): boolean {
  function relativeLuminance(hex: string): number {
    const clean = hex.replace('#', '');
    const r = parseInt(clean.slice(0, 2), 16) / 255;
    const g = parseInt(clean.slice(2, 4), 16) / 255;
    const b = parseInt(clean.slice(4, 6), 16) / 255;
    const linearise = (c: number) =>
      c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
    return 0.2126 * linearise(r) + 0.7152 * linearise(g) + 0.0722 * linearise(b);
  }
  const la = relativeLuminance(hexA);
  const lb = relativeLuminance(hexB);
  const lighter = Math.max(la, lb);
  const darker = Math.min(la, lb);
  const ratio = (lighter + 0.05) / (darker + 0.05);
  return ratio >= 4.5;
}
