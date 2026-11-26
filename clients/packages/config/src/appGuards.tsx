import React, { useEffect, useState } from 'react';
import { posthog } from '@voxire/observability';
import { FLAG_KEYS, FLAG_PROD_DEFAULTS, FLAG_ENV_OVERRIDES, type FlagKey } from './flags';
import { resolveApiBase } from './serviceUrls';

/**
 * Client-side feature-flag resolver.
 *
 * Order:
 *   1. Server-side Super-Admin flag from `POST /hub/feature/v1/features`
 *      (authoritative platform kill-switch matching the Go backend)
 *   2. PostHog flag value (live, overridable per cohort)
 *   3. VITE_APP_*_ENABLED env var from the build
 *   4. Hard-coded prod default.
 *
 * Both server flags and PostHog load asynchronously on browser mount;
 * `useFeatureFlag` falls through to env/default immediately and
 * re-evaluates as soon as either source updates.
 */

// Local env accessor - keeps consumers' tsconfigs free of a
// `vite/client` dependency for what is just an optional lookup.
const viteEnv = (import.meta as unknown as { env?: Record<string, string | undefined> }).env ?? {};

let serverFlags: Record<string, boolean> | null = null;
let serverFlagsPromise: Promise<void> | null = null;
const serverFlagListeners = new Set<() => void>();

function notifyServerFlagListeners(): void {
  serverFlagListeners.forEach((cb) => {
    try {
      cb();
    } catch {
      // Ignore individual subscriber errors.
    }
  });
}

function isTestRuntime(): boolean {
  return viteEnv.MODE === 'test' || Boolean(viteEnv.VITEST);
}

/**
 * Fetches public Super-Admin feature flags from `/hub/feature/v1/features`
 * once per session and notifies all mounted `useFeatureFlag` subscribers.
 * Fails open silently if offline or unreachable.
 */
export function hydrateServerFlags(customFetch?: typeof fetch): Promise<void> {
  if (serverFlagsPromise) return serverFlagsPromise;
  const fetchImpl = customFetch ?? (typeof globalThis.fetch === 'function' ? globalThis.fetch : undefined);
  if (!fetchImpl) return Promise.resolve();

  const endpoint = `${resolveApiBase()}/hub/feature/v1/features`;
  serverFlagsPromise = fetchImpl(endpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ subject: 'web', attributes: {} }),
  })
    .then(async (res) => {
      if (!res.ok) return;
      const payload = (await res.json()) as {
        features?: Record<string, { enabled?: unknown } | undefined>;
      };
      const raw = payload?.features;
      if (!raw || typeof raw !== 'object') return;
      const resolved: Record<string, boolean> = {};
      for (const [key, entry] of Object.entries(raw)) {
        if (entry && typeof entry.enabled === 'boolean') {
          resolved[key] = entry.enabled;
        }
      }
      serverFlags = resolved;
      notifyServerFlagListeners();
    })
    .catch(() => {
      // Fail open on network/parse errors so offline/degraded API never breaks UI.
    });
  return serverFlagsPromise;
}

export function resetServerFlagsForTest(): void {
  serverFlags = null;
  serverFlagsPromise = null;
}

function readServerFlag(flag: FlagKey): boolean | null {
  if (serverFlags !== null && typeof serverFlags[flag] === 'boolean') {
    return serverFlags[flag];
  }
  return null;
}

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
  const server = readServerFlag(flag);
  if (server !== null) return server;
  const ph = readPostHog(flag);
  if (ph !== null) return ph;
  const env = readEnvOverride(flag);
  if (env !== null) return env;
  return FLAG_PROD_DEFAULTS[flag];
}

/**
 * React hook - re-renders when Super-Admin server flags or PostHog
 * report new flag values. Use this inside components. For route-level
 * gating prefer `<AppFlagGuard>` below, which handles the fallback UI.
 */
export function useFeatureFlag(flag: FlagKey): boolean {
  const [enabled, setEnabled] = useState<boolean>(() => resolveFlag(flag));

  useEffect(() => {
    setEnabled(resolveFlag(flag));
    const onUpdate = () => setEnabled(resolveFlag(flag));
    serverFlagListeners.add(onUpdate);
    if (typeof window !== 'undefined' && !isTestRuntime()) {
      void hydrateServerFlags();
    }

    let unsub: (() => void) | undefined;
    try {
      unsub = posthog.onFeatureFlags?.(onUpdate);
    } catch {
      // PostHog not initialised - fall through, server/env/default already applied.
    }
    return () => {
      serverFlagListeners.delete(onUpdate);
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
