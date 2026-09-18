import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { bootObservability } from './boot';

const { initSentryMock, initPostHogMock, disableMock, replayMock } = vi.hoisted(() => ({
  initSentryMock: vi.fn(), initPostHogMock: vi.fn(), disableMock: vi.fn(), replayMock: vi.fn(),
}));
vi.mock('./sentry', () => ({ initSentry: initSentryMock, enableSentryReplays: replayMock }));
vi.mock('./posthog', () => ({ initPostHog: initPostHogMock, disablePostHog: disableMock }));

let cleanup: (() => void) | undefined;
function decide(state: string): void {
  document.cookie = `rtylr-consent=${state}; path=/`;
  window.dispatchEvent(new CustomEvent('rtylr-consent-change', { detail: state }));
}
beforeEach(() => { vi.clearAllMocks(); document.cookie = 'rtylr-consent=; Max-Age=0; path=/'; });
afterEach(() => { cleanup?.(); cleanup = undefined; });

describe('analytics consent lifecycle', () => {
  it('keeps optional analytics off until acceptance', () => {
    cleanup = bootObservability({ app: 'auth' });
    expect(initSentryMock).toHaveBeenCalledOnce();
    expect(initPostHogMock).not.toHaveBeenCalled();
    decide('denied');
    expect(initPostHogMock).not.toHaveBeenCalled();
    decide('granted');
    expect(initPostHogMock).toHaveBeenCalledWith({ app: 'auth' });
    expect(replayMock).not.toHaveBeenCalled();
  });
  it('stops after withdrawal and allows a later fresh grant', () => {
    decide('granted');
    cleanup = bootObservability({ app: 'hr' });
    decide('undecided');
    expect(disableMock).toHaveBeenCalledOnce();
    decide('denied');
    expect(disableMock).toHaveBeenCalledTimes(2);
    decide('granted');
    expect(initPostHogMock).toHaveBeenCalledTimes(2);
  });
  it('does not trust a grant event without a persisted choice', () => {
    cleanup = bootObservability({ app: 'recruit' });
    window.dispatchEvent(new CustomEvent('rtylr-consent-change', { detail: 'granted' }));
    expect(initPostHogMock).not.toHaveBeenCalled();
  });
  it('reconciles choices from another tab on focus and removes listeners on cleanup', () => {
    decide('granted');
    cleanup = bootObservability({ app: 'crm' });
    document.cookie = 'rtylr-consent=denied; path=/';
    window.dispatchEvent(new Event('focus'));
    expect(disableMock).toHaveBeenCalledOnce();
    cleanup();
    decide('granted');
    expect(initPostHogMock).toHaveBeenCalledOnce();
  });
});
