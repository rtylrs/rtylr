import React, { useEffect, useState } from 'react';
import { posthog } from '@voxire/observability';
import { FLAG_KEYS, FLAG_PROD_DEFAULTS, FLAG_ENV_OVERRIDES, type FlagKey } from './flags';
import { allFeaturesEnabled } from './runtimeConfig';

/**
 * Client-side feature-flag resolver.
 *
 * Order (documented in `flags.ts`):
 *   1. PostHog flag value (live, overridable per cohort)
 *   2. VITE_APP_*_ENABLED env var from the build
 *   3. Hard-coded prod default.
 *
 * PostHog may not have loaded yet when the first render runs - in
 * that case we fall through to env/default and re-check once PostHog
 * fires its `onFeatureFlags` callback.
 */

// Local env accessor - keeps consumers' tsconfigs free of a
// `vite/client` dependency for what is just an optional lookup.
const viteEnv = (import.meta as unknown as { env?: Record<string, string | undefined> }).env ?? {};

function readEnvOverride(flag: FlagKey): boolean | null {
  const envName = FLAG_ENV_OVERRIDES[flag];
  if (!envName) return null;
  const raw = viteEnv[envName];
  if (raw === '1' || raw === 'true') return true;
  if (raw === '0' || raw === 'false') return false;
  return null;
}

function readPostHog(flag: FlagKey): boolean | null {
  // posthog-js returns `undefined` when flags haven't loaded, a
  // boolean when the flag exists, or a string (multi-variant) that
  // we treat as truthy except 'off'/'false'.
  try {
    const value = posthog.isFeatureEnabled?.(flag);
    if (typeof value === 'boolean') return value;
  } catch {
    // posthog not initialised yet - consent probably wasn't granted.
  }
  return null;
}

export function resolveFlag(flag: FlagKey): boolean {
  // Self-host installs run without the hosted licensing/upsell gating: every
  // app and feature is treated as enabled. This is centralized here so every
  // `useFeatureFlag` / `<AppFlagGuard>` consumer inherits it. Hosted (and any
  // deployment where runtime config didn't flip this on) is unaffected.
  if (allFeaturesEnabled()) return true;
  const ph = readPostHog(flag);
  if (ph !== null) return ph;
  const env = readEnvOverride(flag);
  if (env !== null) return env;
  return FLAG_PROD_DEFAULTS[flag];
}

/**
 * React hook - re-renders when PostHog reports new flag values. Use
 * this inside components. For route-level gating prefer
 * `<AppFlagGuard>` below, which handles the fallback UI.
 */
export function useFeatureFlag(flag: FlagKey): boolean {
  const [enabled, setEnabled] = useState<boolean>(() => resolveFlag(flag));

  useEffect(() => {
    // Re-resolve once PostHog publishes flags. The callback fires
    // every time the feature-flag map changes, so one subscription
    // handles both the "flags just loaded" case and the "admin
    // flipped a flag in PostHog" case.
    let unsub: (() => void) | undefined;
    try {
      unsub = posthog.onFeatureFlags?.(() => setEnabled(resolveFlag(flag)));
    } catch {
      // PostHog not initialised - fall through, env/default already
      // applied.
    }
    return () => {
      if (unsub) unsub();
    };
  }, [flag]);

  return enabled;
}

interface AppFlagGuardProps {
  flag: FlagKey;
  children: React.ReactNode;
  /**
   * Shown when the flag evaluates to false. Defaults to a simple
   * "Coming soon" screen. Pass `null` to render nothing (useful for
   * sub-feature gates that should just hide a button).
   */
  fallback?: React.ReactNode;
}

const DEFAULT_FALLBACK: React.ReactNode = (
  <div
    style={{
      display: 'flex',
      minHeight: '60vh',
      alignItems: 'center',
      justifyContent: 'center',
      textAlign: 'center',
      padding: '48px 24px',
      fontFamily: 'system-ui, -apple-system, sans-serif',
      color: '#333',
    }}
  >
    <div>
      <h1 style={{ fontSize: '1.5rem', fontWeight: 600, marginBottom: 12 }}>Coming soon</h1>
      <p style={{ color: '#6b6b6b', fontSize: '0.95rem', maxWidth: 420 }}>
        This rtylr app isn&rsquo;t live yet. We&rsquo;ll light it up the moment it passes our own
        launch gate.
      </p>
    </div>
  </div>
);

/**
 * Gate a whole app (or a whole route subtree) behind a flag. Wrap
 * the app's top-level routes in this - typical usage in `App.tsx`:
 *
 * ```tsx
 * import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
 *
 * const App = () => (
 *   <AppFlagGuard flag={FLAG_KEYS.APP_ERP}>
 *     <Routes>...</Routes>
 *   </AppFlagGuard>
 * );
 * ```
 */
export const AppFlagGuard: React.FC<AppFlagGuardProps> = ({
  flag,
  children,
  fallback = DEFAULT_FALLBACK,
}): React.JSX.Element => {
  const enabled = useFeatureFlag(flag);
  if (!enabled) return <>{fallback}</>;
  return <>{children}</>;
};
