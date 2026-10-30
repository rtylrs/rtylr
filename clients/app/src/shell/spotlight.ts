const SPOTLIGHT_CLASS = 'portal-spotlight';
const CONTROLS = 'button, a[href], [role="button"], [role="tab"], [role="menuitem"]';

function normalize(text: string): string {
  return text.toLowerCase().replace(/[+…]/g, ' ').replace(/\s+/g, ' ').trim();
}

function findControl(root: ParentNode, label: string): HTMLElement | null {
  const wanted = normalize(label);
  if (!wanted) return null;
  const visible = Array.from(root.querySelectorAll<HTMLElement>(CONTROLS)).filter(
    (element) => element.getClientRects().length > 0 && !element.closest('[aria-hidden="true"]')
  );
  const text = (element: HTMLElement) =>
    normalize(element.getAttribute('aria-label') || element.textContent || '');
  return (
    visible.find((element) => text(element) === wanted) ??
    (wanted.length >= 4 ? visible.find((element) => text(element).includes(wanted)) : undefined) ??
    null
  );
}

/**
 * Waits for the page to show a control labelled `label`, then scrolls to it
 * and highlights it for a moment, so the person sees which button Caret
 * meant. Gives up quietly when the page has no such control. Returns a
 * cancel function.
 */
export function spotlightControl(
  label: string,
  root: () => ParentNode | null = () => document.getElementById('main-content'),
  timeoutMs = 4000
): () => void {
  const started = Date.now();
  let timer = 0;
  const tick = () => {
    const container = root();
    const control = container && findControl(container, label);
    if (control) {
      const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches;
      control.scrollIntoView?.({ block: 'center', behavior: reduced ? 'auto' : 'smooth' });
      control.classList.add(SPOTLIGHT_CLASS);
      window.setTimeout(() => control.classList.remove(SPOTLIGHT_CLASS), 3400);
      return;
    }
    if (Date.now() - started < timeoutMs) timer = window.setTimeout(tick, 150);
  };
  timer = window.setTimeout(tick, 100);
  return () => window.clearTimeout(timer);
}
