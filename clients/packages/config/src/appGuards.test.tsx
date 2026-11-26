import { afterEach, describe, expect, test, vi } from 'vitest';
import { render, screen, cleanup, act } from '@testing-library/react';
import React from 'react';
import { AppFlagGuard, hydrateServerFlags, resetServerFlagsForTest } from './appGuards';
import { FLAG_KEYS, FLAG_PROD_DEFAULTS } from './flags';

/**
 * AppFlagGuard tests. PostHog is mocked to be absent so all paths
 * fall through to the env-override / prod-default fallback chain.
 *
 * Note: env mutation between tests is unreliable in Vitest because
 * appGuards reads `import.meta.env` at module load. The tests below
 * exercise the prod-default branch only, which is stable. Env-
 * override behaviour is covered by the integration tests in each
 * app's own typecheck (the flag values surface in the UI).
 */

vi.mock('@voxire/observability', () => ({
  posthog: {
    isFeatureEnabled: undefined,
    onFeatureFlags: undefined,
  },
}));

afterEach(() => {
  cleanup();
  resetServerFlagsForTest();
});

describe('AppFlagGuard', () => {
  test('renders children when prod default for the flag is true', () => {
    // Find a flag whose prod default is `true`. Recruit is on by
    // default per the plan-locked rollout.
    const enabledFlag = (
      Object.keys(FLAG_PROD_DEFAULTS) as Array<keyof typeof FLAG_PROD_DEFAULTS>
    ).find((k) => FLAG_PROD_DEFAULTS[k] === true);
    expect(enabledFlag).toBeDefined();

    render(
      <AppFlagGuard flag={enabledFlag!} fallback={<span>blocked</span>}>
        <span>protected</span>
      </AppFlagGuard>
    );
    expect(screen.queryByText('protected')).not.toBeNull();
    expect(screen.queryByText('blocked')).toBeNull();
  });

  test('renders fallback when prod default for the flag is false', () => {
    // Find a flag whose prod default is `false` - most app flags are
    // off until launch gate flips them.
    const disabledFlag = (
      Object.keys(FLAG_PROD_DEFAULTS) as Array<keyof typeof FLAG_PROD_DEFAULTS>
    ).find((k) => FLAG_PROD_DEFAULTS[k] === false);
    expect(disabledFlag).toBeDefined();

    render(
      <AppFlagGuard flag={disabledFlag!} fallback={<span>blocked</span>}>
        <span>protected</span>
      </AppFlagGuard>
    );
    expect(screen.queryByText('blocked')).not.toBeNull();
    expect(screen.queryByText('protected')).toBeNull();
  });

  test('FLAG_KEYS surface stable string values', () => {
    expect(FLAG_KEYS.APP_RECRUIT).toBe('app.recruit');
    expect(FLAG_KEYS.APP_ERP).toBe('app.erp');
  });

  test('Super-Admin server flag hydration overrides prod default live', async () => {
    render(
      <AppFlagGuard flag={FLAG_KEYS.APP_MENU_ORDERING} fallback={<span>disabled-by-admin</span>}>
        <span>ordering-active</span>
      </AppFlagGuard>
    );
    expect(screen.queryByText('ordering-active')).not.toBeNull();

    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        features: {
          [FLAG_KEYS.APP_MENU_ORDERING]: { key: FLAG_KEYS.APP_MENU_ORDERING, enabled: false },
        },
      }),
    }) as unknown as typeof fetch;

    await act(async () => {
      await hydrateServerFlags(mockFetch);
    });

    expect(screen.queryByText('disabled-by-admin')).not.toBeNull();
    expect(screen.queryByText('ordering-active')).toBeNull();
  });

  test('server flag hydration fails open on network errors', async () => {
    const failingFetch = vi.fn().mockRejectedValue(new Error('offline')) as unknown as typeof fetch;

    await act(async () => {
      await hydrateServerFlags(failingFetch);
    });

    render(
      <AppFlagGuard flag={FLAG_KEYS.APP_MENU_ORDERING} fallback={<span>disabled-by-admin</span>}>
        <span>ordering-active</span>
      </AppFlagGuard>
    );
    expect(screen.queryByText('ordering-active')).not.toBeNull();
  });
});
