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
const viteEnv = (import.meta as { env?: Record<string, string | undefined> }).env ?? {};

export function initPostHog(opts: InitPostHogOptions): void {
  if (initialised) return;
  const key = (viteEnv.VITE_POSTHOG_KEY as string | undefined) ?? '';
  if (!key) return;

  posthog.init(key, {
    // US region - confirmed region for this project. Changing
    // this means re-migrating every existing event, so it's nailed
    // here rather than read from env.
    api_host: 'https://us.i.posthog.com',
    ui_host: 'https://us.posthog.com',
    // We handle consent upstream (strict opt-in). When PostHog
    // sees `opt_out_persistence_by_default: true`, it won't write
    // its persistence cookie until we explicitly opt in.
    opt_out_persistence_by_default: false,
    // Autocapture captures clicks / page views automatically -
    // good default for a first-week learning loop. Turn off later
    // once we have explicit tracking per feature.
    autocapture: true,
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
