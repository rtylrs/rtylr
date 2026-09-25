import posthog from 'posthog-js';

/**
 * PostHog initialisation for every rtylr frontend app. US region.
 *
 * Gated on consent: this is only called by `@voxire/consent` after
 * the user has accepted cookies. Never call `initPostHog` directly
 * from an app's main.tsx - route through the consent package.
 */
interface InitPostHogOptions {
  app: string;
}

let initialised = false;

// See sentry.ts - local accessor avoids needing vite/client types
// downstream.
const viteEnv = (import.meta as unknown as { env?: Record<string, string | undefined> }).env ?? {};

export function initPostHog(opts: InitPostHogOptions): void {
  if (initialised) {
    posthog.opt_in_capturing({ captureEventName: false });
    return;
  }
  const key = (viteEnv.VITE_POSTHOG_KEY as string | undefined) ?? '';
  if (!key) return;

  posthog.init(key, {
    // US region - confirmed region for this project. Changing
    // this means re-migrating every existing event, so it's nailed
    // here rather than read from env.
    api_host: 'https://us.i.posthog.com',
    ui_host: 'https://us.posthog.com',
    // Withdrawal clears persistence as well as stopping capture.
    opt_out_persistence_by_default: true,
    // HR, payroll, and recruitment screens can contain sensitive data.
    // Generic analytics consent is not permission to capture field content.
    autocapture: false,
    capture_pageview: true,
    capture_pageleave: true,
    // Tag every event with the originating app so we can split
    // funnels per subdomain.
    loaded: (ph) => {
      ph.register({ app: opts.app });
    },
    // Disable session recording by default - it's the single most
    // expensive and PII-dense surface. Turn it on per cohort via
    // PostHog cloud if we need it.
    disable_session_recording: true,
  });
  initialised = true;
  posthog.opt_in_capturing({ captureEventName: false });
}

export function disablePostHog(): void {
  if (!initialised) return;
  posthog.opt_out_capturing();
}

export function identifyUser(userId: string, traits?: Record<string, unknown>): void {
  if (!initialised) return;
  posthog.identify(userId, traits);
}

export function captureEvent(event: string, props?: Record<string, unknown>): void {
  if (!initialised) return;
  posthog.capture(event, props);
}

export function resetAnalytics(): void {
  if (!initialised) return;
  posthog.reset();
}

export { posthog };
