import * as Sentry from '@sentry/react';

/**
 * Sentry initialisation shared across every rtylr frontend app.
 *
 * Reads `VITE_SENTRY_DSN` from env. A missing DSN is not an error -
 * local dev and preview builds run without it. Calling `initSentry`
 * twice is a no-op thanks to Sentry's internal guard.
 *
 * Consent: we initialise Sentry eagerly (for error capture) but keep
 * `replays` and session tracking OFF until the user accepts cookies.
 * Generic analytics consent does not enable replays. A separately reviewed
 * recording feature would need its own disclosure and consent lifecycle.
 */
interface InitSentryOptions {
  /** App identifier ("flow", "dash", etc.). Surfaces in Sentry `tags.app`. */
  app: string;
  /** Release tag - typically the git SHA. Groups events across deploys. */
  release?: string;
}

let initialised = false;

// Type-safe accessor for `import.meta.env` without pulling the whole
// `vite/client` types into every transitive consumer's tsconfig.
const viteEnv = (import.meta as { env?: Record<string, string | boolean | undefined> }).env ?? {};

export function initSentry(opts: InitSentryOptions): void {
  if (initialised) return;
  const dsn = (viteEnv.VITE_SENTRY_DSN as string | undefined) ?? '';
  if (!dsn) return;

  const environment =
    (viteEnv.VITE_APP_ENV as string | undefined) ?? (viteEnv.PROD ? 'production' : 'development');

  Sentry.init({
    dsn,
    environment,
    release: opts.release,
    // Performance tracing at 10 % - enough to spot regressions without
    // burning the quota on every page view.
    tracesSampleRate: 0.1,
    // Session Replay stays off in the shared boot path.
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 0,
    // Strip common noisy errors (ResizeObserver, network aborts) so
    // the dashboard stays signal-rich.
    ignoreErrors: [
      'ResizeObserver loop limit exceeded',
      'ResizeObserver loop completed with undelivered notifications.',
      'Non-Error promise rejection captured',
      'Load failed',
      'NetworkError',
    ],
    beforeSend(event) {
      // Tag every event with the calling app so the dashboard can
      // filter errors by subdomain.
      event.tags = { ...(event.tags ?? {}), app: opts.app };
      return event;
    },
  });
  initialised = true;
}

/**
 * Call from the consent package when the user accepts cookies -
 * flips replays on for session recording (on error only; full-session
 * recording is too invasive and PII-heavy for a B2B SaaS).
 */
export function enableSentryReplays(): void {
  if (!initialised) return;
  Sentry.addIntegration(
    Sentry.replayIntegration({
      maskAllText: true,
      blockAllMedia: true,
    })
  );
}

/**
 * Re-exports the raw Sentry namespace for components that need to
 * call `Sentry.captureException` or set scope tags directly.
 */
export { Sentry };
