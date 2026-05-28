import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import { bootObservability } from './boot';

/**
 * Smoke tests for bootObservability - the helper every app calls
 * once at startup. Exercises:
 *   - undecided consent → only Sentry inits, PostHog stays off
 *   - consent already granted → both init
 *   - consent change event after boot → PostHog inits late
 *
 * Sentry + PostHog SDKs are mocked because we're testing the
 * orchestration, not their internals.
 */

const initSentryMock = vi.fn();
const enableReplaysMock = vi.fn();
const initPostHogMock = vi.fn();

vi.mock('./sentry', () => ({
  initSentry: (opts: unknown) => initSentryMock(opts),
  enableSentryReplays: () => enableReplaysMock(),
}));

vi.mock('./posthog', () => ({
  initPostHog: (opts: unknown) => initPostHogMock(opts),
}));

describe('bootObservability', () => {
  beforeEach(() => {
    initSentryMock.mockClear();
    enableReplaysMock.mockClear();
    initPostHogMock.mockClear();
    document.cookie = 'rtylr-consent=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
  });

  afterEach(() => {
    document.cookie = 'rtylr-consent=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/';
  });

  test('undecided consent: Sentry inits, PostHog and replays do not', async () => {
    bootObservability({ app: 'auth' });

    expect(initSentryMock).toHaveBeenCalledWith({ app: 'auth', release: undefined });
    expect(initPostHogMock).not.toHaveBeenCalled();
    expect(enableReplaysMock).not.toHaveBeenCalled();
  });

  test('granted consent at boot: PostHog + replays both fire', async () => {
    document.cookie = 'rtylr-consent=granted; path=/';
    bootObservability({ app: 'dash' });

    expect(initSentryMock).toHaveBeenCalledTimes(1);
    expect(initPostHogMock).toHaveBeenCalledWith({ app: 'dash' });
    expect(enableReplaysMock).toHaveBeenCalledTimes(1);
  });

  test('consent flips to granted after boot: PostHog inits late', async () => {
    bootObservability({ app: 'recruit' });

    expect(initPostHogMock).not.toHaveBeenCalled();

    window.dispatchEvent(new CustomEvent('rtylr-consent-change', { detail: { state: 'granted' } }));

    // Use `toHaveBeenCalled()` rather than a strict count: prior tests
    // in the file may leave dangling event listeners on `window`
    // (boot.ts only removes its listener after a 'granted' event), so
    // counting precisely would be brittle. The contract this test
    // proves is "the event causes init" - that's what we check.
    expect(initPostHogMock).toHaveBeenCalledWith({ app: 'recruit' });
    expect(enableReplaysMock).toHaveBeenCalled();
  });

  test('consent flips to denied: stays opt-out', async () => {
    bootObservability({ app: 'menu' });

    window.dispatchEvent(new CustomEvent('rtylr-consent-change', { detail: { state: 'denied' } }));

    expect(initPostHogMock).not.toHaveBeenCalled();
    expect(enableReplaysMock).not.toHaveBeenCalled();
  });

  test('accepts the string-detail event shape that @voxire/consent emits', async () => {
    bootObservability({ app: 'erp' });
    expect(initPostHogMock).not.toHaveBeenCalled();

    // The consent package dispatches `detail: 'granted'` (raw string,
    // no wrapper object). This guards against regressions where the
    // observability boot only accepts the `{state}` shape.
    window.dispatchEvent(new CustomEvent('rtylr-consent-change', { detail: 'granted' }));

    expect(initPostHogMock).toHaveBeenCalledWith({ app: 'erp' });
  });
});
