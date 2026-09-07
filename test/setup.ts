import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

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

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

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
