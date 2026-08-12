import { describe, it, expect, vi, beforeEach } from 'vitest';
import { initSentry, enableSentryReplays } from './sentry';
import { initPostHog, identifyUser, captureEvent, resetAnalytics } from './posthog';

describe('Observability package', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('no-ops initSentry when VITE_SENTRY_DSN is not set', () => {
    expect(() => initSentry({ app: 'dash' })).not.toThrow();
  });

  it('no-ops enableSentryReplays when uninitialised', () => {
    expect(() => enableSentryReplays()).not.toThrow();
  });

  it('no-ops initPostHog when VITE_POSTHOG_KEY is not set', () => {
    expect(() => initPostHog({ app: 'dash' })).not.toThrow();
  });

  it('no-ops tracking functions when posthog is uninitialised', () => {
    expect(() => identifyUser('user-123')).not.toThrow();
    expect(() => captureEvent('button_clicked')).not.toThrow();
    expect(() => resetAnalytics()).not.toThrow();
  });
});
