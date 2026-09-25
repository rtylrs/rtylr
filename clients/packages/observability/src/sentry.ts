import * as Sentry from "@sentry/react";

/**
 * Sentry initialisation shared across every rtylr frontend app.
 *
 * Reads `VITE_SENTRY_DSN` from env. A missing DSN is not an error -
 * local dev and preview builds run without it. Calling `initSentry`
 * twice is a no-op thanks to Sentry's internal guard.
 *
 * Consent: we initialise Sentry eagerly for error capture, but keep release
 * sessions, performance tracing, and replays off until the user accepts
 * cookies. Withdrawal returns those optional signals to zero without disabling
 * error reporting. Generic analytics consent does not enable replays; a
 * separately reviewed recording feature still needs its own consent lifecycle.
 */
interface InitSentryOptions {
  /** App identifier ("flow", "dash", etc.). Surfaces in Sentry `tags.app`. */
  app: string;
  /** Release tag - typically the git SHA. Groups events across deploys. */
  release?: string;
}

let initialised = false;
let consentGranted = false;

const CONSENT_COOKIE = "rtylr-consent";
const CONSENT_EVENT = "rtylr-consent-change";

// Type-safe accessor for `import.meta.env` without pulling the whole
// `vite/client` types into every transitive consumer's tsconfig.
const viteEnv =
  (import.meta as unknown as { env?: Record<string, string | boolean | undefined> }).env ??
  {};

function hasConsent(): boolean {
  if (typeof document === "undefined") return false;
  return document.cookie
    .split(";")
    .some((part) => part.trim() === `${CONSENT_COOKIE}=granted`);
}

function syncSentryConsent(): void {
  const nextConsentGranted = hasConsent();
  if (nextConsentGranted === consentGranted) return;
  consentGranted = nextConsentGranted;

  if (!consentGranted) {
    Sentry.endSession();
    return;
  }

  if (!Sentry.getClient()?.getIntegrationByName("BrowserSession")) {
    Sentry.addIntegration(Sentry.browserSessionIntegration());
    return;
  }

  // The integration can be installed only once. Start a fresh release-health
  // session when a user grants consent again after withdrawal.
  Sentry.startSession({ ignoreDuration: true });
  Sentry.captureSession();
}

export function createSentryOptions(
  opts: InitSentryOptions,
  dsn: string,
  environment: string,
  isConsented: () => boolean = () => consentGranted,
): Parameters<typeof Sentry.init>[0] {
  return {
    dsn,
    environment,
    release: opts.release,
    integrations: (integrations) =>
      isConsented()
        ? integrations
        : integrations.filter(
            (integration) => integration.name !== "BrowserSession",
          ),
    // Performance tracing is sampled at 10% only after opt-in. A sampler keeps
    // withdrawal effective without reinitialising the error client.
    tracesSampler: () => (isConsented() ? 0.1 : 0),
    // BrowserSession cannot be removed after installation, so keep a final gate
    // for sessions queued at or after withdrawal.
    transport: (transportOptions) => {
      const transport = Sentry.makeFetchTransport(transportOptions);
      return {
        ...transport,
        send: (envelope) => {
          const isOptionalTelemetry = envelope[1].some(
            ([item]) =>
              item.type === "session" ||
              item.type === "sessions" ||
              item.type === "transaction",
          );
          if (!isConsented() && isOptionalTelemetry) {
            return Promise.resolve({ statusCode: 200 });
          }
          return transport.send(envelope);
        },
      };
    },
    // Session Replay stays off in the shared boot path.
    replaysSessionSampleRate: 0,
    replaysOnErrorSampleRate: 0,
    // Strip common noisy errors (ResizeObserver, network aborts) so
    // the dashboard stays signal-rich.
    ignoreErrors: [
      "ResizeObserver loop limit exceeded",
      "ResizeObserver loop completed with undelivered notifications.",
      "Non-Error promise rejection captured",
      "Load failed",
      "NetworkError",
    ],
    beforeSend(event) {
      // Tag every event with the calling app so the dashboard can
      // filter errors by subdomain.
      event.tags = { ...(event.tags ?? {}), app: opts.app };
      return event;
    },
    beforeSendTransaction(event) {
      if (!isConsented()) return null;
      event.tags = { ...(event.tags ?? {}), app: opts.app };
      return event;
    },
  };
}

export function initSentry(opts: InitSentryOptions): void {
  if (initialised) return;
  const dsn = (viteEnv.VITE_SENTRY_DSN as string | undefined) ?? "";
  if (!dsn) return;

  const environment =
    (viteEnv.VITE_APP_ENV as string | undefined) ??
    (viteEnv.PROD ? "production" : "development");

  consentGranted = hasConsent();

  Sentry.init(createSentryOptions(opts, dsn, environment));
  initialised = true;

  if (typeof window !== "undefined") {
    window.addEventListener(CONSENT_EVENT, syncSentryConsent);
    window.addEventListener("focus", syncSentryConsent);
  }
}

/**
 * Call from the consent package when the user accepts cookies -
 * flips replays on for session recording (on error only; full-session
 * recording is too invasive and PII-heavy for a B2B SaaS).
 */
export function enableSentryReplays(): void {
  if (!initialised || !consentGranted) return;
  Sentry.addIntegration(
    Sentry.replayIntegration({
      maskAllText: true,
      blockAllMedia: true,
    }),
  );
}

/**
 * Re-exports the raw Sentry namespace for components that need to
 * call `Sentry.captureException` or set scope tags directly.
 */
export { Sentry };
