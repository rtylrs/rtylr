/**
 * A `matchMedia` that evaluates, plus viewport control for tests.
 *
 * Measured, because the two environments differ and the difference matters:
 *
 * - **happy-dom 20** ships a working `matchMedia`. It evaluates width features
 *   against its own viewport and re-evaluates after that viewport changes. It
 *   does not fire the `change` event.
 * - **jsdom 29** does not implement `matchMedia` at all. Calling it throws
 *   `window.matchMedia is not a function`.
 *
 * So `setup.ts` needed a shim for jsdom, and had one. Its flaw was not that it
 * existed but that it reported `matches: false` for **every** query, which
 * disabled every responsive branch in the app: a component choosing a layout
 * from a media query always rendered its no-match side, and no assertion could
 * tell that apart from the viewport genuinely being narrow.
 *
 * This replaces it with one that evaluates `min-width` and `max-width` against
 * `window.innerWidth`. Anything else still returns false, so
 * `(prefers-color-scheme: dark)` and the rest behave exactly as before.
 *
 * Installed for both environments rather than only where it is missing, so a
 * test does not silently depend on which DOM it happens to run under.
 *
 * rtylr#1215 is what needed this: the calendar drops its month grid for a list
 * below 768px, and that is only observable through a media query.
 */

const WIDTH_FEATURE = /\(\s*(min|max)-width\s*:\s*(\d+(?:\.\d+)?)px\s*\)/gi;

type Listener = (event: MediaQueryListEvent) => void;

interface Entry {
  query: string;
  listeners: Set<Listener>;
  last: boolean;
}

const entries = new Set<Entry>();
let width = 1024;

/** The width the design work targets as its floor. */
export const MOBILE_WIDTH = 390;
/** The default both test DOMs start at. */
export const DESKTOP_WIDTH = 1024;

/** Evaluates only the width features. Anything else is unknown, so false. */
export function evaluateQuery(query: string, at: number): boolean {
  const clauses = [...query.matchAll(WIDTH_FEATURE)];
  if (clauses.length === 0) return false;
  return clauses.every(([, kind, px]) =>
    kind.toLowerCase() === 'min' ? at >= Number(px) : at <= Number(px)
  );
}

export function installMatchMedia(): void {
  width = window.innerWidth || DESKTOP_WIDTH;
  entries.clear();

  window.matchMedia = ((query: string) => {
    const entry: Entry = { query, listeners: new Set(), last: evaluateQuery(query, width) };
    entries.add(entry);

    return {
      get matches() {
        return evaluateQuery(query, width);
      },
      media: query,
      onchange: null,
      addEventListener: (_type: string, listener: Listener) => entry.listeners.add(listener),
      removeEventListener: (_type: string, listener: Listener) => entry.listeners.delete(listener),
      // The deprecated pair, in case a dependency still reaches for it.
      addListener: (listener: Listener) => entry.listeners.add(listener),
      removeListener: (listener: Listener) => entry.listeners.delete(listener),
      dispatchEvent: () => false,
    } as unknown as MediaQueryList;
  }) as typeof window.matchMedia;
}

/**
 * Set the viewport width and notify listeners, which is more than either
 * environment does on its own: neither fires `change`, so without this a
 * component would only ever see the width it first rendered at.
 */
export function setViewportWidth(next: number): void {
  width = next;
  Object.defineProperty(window, 'innerWidth', {
    value: next,
    writable: true,
    configurable: true,
  });

  for (const entry of entries) {
    const matches = evaluateQuery(entry.query, next);
    if (matches === entry.last) continue;
    entry.last = matches;
    for (const listener of entry.listeners) {
      listener({ matches, media: entry.query } as MediaQueryListEvent);
    }
  }
  window.dispatchEvent(new Event('resize'));
}

export function resetViewport(): void {
  setViewportWidth(DESKTOP_WIDTH);
}
