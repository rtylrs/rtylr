/**
 * happy-dom and color-mix
 *
 * The suite runs on happy-dom, whose CSSStyleDeclaration **silently discards any
 * `color-mix()` value assigned through the style property**, which is the path
 * React uses. `withAlpha` produces exactly that form and Flow calls it 31 times,
 * so every tint in the app is absent from the DOM in this environment. Nothing
 * warns; the declaration simply is not there.
 *
 * Measured (rtylr#1249):
 *
 *   value                                     setAttribute   el.style.x
 *   #112233                                   kept           kept
 *   color-mix(in srgb, #112233 13%, …)        kept           DROPPED
 *   color-mix(… var(--flow-text))             kept           DROPPED
 *   var(--flow-text)                          kept           kept
 *
 * jsdom keeps all four. So a test that needs to see a tint, or a foreground
 * derived from one, opens with:
 *
 *   // @vitest-environment jsdom
 *
 * `Calendar.test.tsx`, `ProjectCalendar.test.tsx` and `TaskCard.test.tsx` do.
 *
 * Why this is worth a comment rather than a footnote: the missing declaration
 * does not make a test fail, it makes an assertion vacuous. Calendar's tint test
 * asserted a hex appeared in the style string and passed for two months on a
 * *sibling* `color: #112233` declaration, which was itself the accessibility
 * defect fixed in rtylr#1246. Removing the defect broke the test, and the break
 * looked like a regression.
 */

import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';
import { installMatchMedia } from './viewport';

// Enable React 19 act environment for testing-library
// @ts-expect-error - IS_REACT_ACT_ENVIRONMENT global
globalThis.IS_REACT_ACT_ENVIRONMENT = true;

// Suppress happy-dom network noise: AbortError on teardown, NetworkError from
// Google Fonts/GTM fetches that can't reach the network in jsdom/happy-dom.
const originalConsoleError = console.error;
console.error = (...args: unknown[]) => {
  const msg = args[0];
  if (msg instanceof DOMException) return;
  if (
    typeof msg === 'string' &&
    /AbortError|NetworkError|fonts\.googleapis|googletagmanager/.test(msg)
  )
    return;
  originalConsoleError(...args);
};

/**
 * No unit test reaches the network.
 *
 * #1204 stopped the suite fetching https://api.voxire.com by pointing the API
 * base at a refused port. That fixed the real problem, a request still in
 * flight when happy-dom tore the window down, but it left the refusal itself
 * surfacing as an unhandled `ECONNREFUSED` in the run output. Failing before a
 * socket is opened is quieter, faster, and says what to do about it.
 *
 * A test that needs `fetch` mocks it, which overrides this.
 */
globalThis.fetch = (() =>
  Promise.reject(
    new Error('Network access is disabled in tests. Mock fetch in the test that needs it.')
  )) as unknown as typeof fetch;

// Cleanup after each test
afterEach(() => {
  cleanup();
});

/**
 * `matchMedia` is shimmed, and the shim evaluates.
 *
 * It has to exist: jsdom 29 does not implement `matchMedia` at all, and calling
 * it throws. happy-dom 20 does implement it. The shim used to report
 * `matches: false` for **every** query, which made every responsive branch in
 * the app untestable, since a component choosing a layout from a media query
 * always rendered its no-match side and nothing could tell that apart from the
 * viewport genuinely being narrow.
 *
 * The replacement evaluates `min-width` and `max-width` and leaves everything
 * else false, so `(prefers-color-scheme: dark)` behaves as it did.
 * `test/viewport.ts` has the measurements and `setViewportWidth`.
 */
installMatchMedia();

// jsdom/happy-dom implement no layout, so Element.prototype.scrollIntoView is
// absent - stub it so components that scroll a ref into view (e.g. an empty
// state's "jump to the form" action) don't throw in tests that click through.
Element.prototype.scrollIntoView = vi.fn();

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] ?? null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => Object.keys(store)[index] ?? null,
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock document.cookie
let cookieStore = '';
Object.defineProperty(document, 'cookie', {
  get: () => cookieStore,
  set: (val: string) => {
    cookieStore = val;
  },
});
