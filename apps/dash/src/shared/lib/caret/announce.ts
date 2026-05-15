/**
 * Speak short status updates through an aria-live region. The region is
 * lazily created on first use so the rest of the app doesn't have to know
 * about it.
 */

let region: HTMLElement | null = null;
let timer: ReturnType<typeof setTimeout> | null = null;

function ensureRegion(): HTMLElement | null {
  if (typeof document === 'undefined') return null;
  if (region) return region;
  region = document.createElement('div');
  region.setAttribute('role', 'status');
  region.setAttribute('aria-live', 'polite');
  region.setAttribute('aria-atomic', 'true');
  region.style.position = 'absolute';
  region.style.width = '1px';
  region.style.height = '1px';
  region.style.overflow = 'hidden';
  region.style.clip = 'rect(0,0,0,0)';
  region.style.whiteSpace = 'nowrap';
  document.body.appendChild(region);
  return region;
}

/** Push a short message to the screen-reader live region. */
export function announce(message: string): void {
  const r = ensureRegion();
  if (!r) return;
  // Some screen readers ignore duplicate text — toggle to force re-read.
  r.textContent = '';
  if (timer) clearTimeout(timer);
  timer = setTimeout(() => {
    if (r) r.textContent = message;
  }, 20);
}
