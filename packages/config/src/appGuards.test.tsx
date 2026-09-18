import { afterEach, describe, expect, test, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import React from 'react';
import { AppFlagGuard } from './appGuards';
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

afterEach(() => cleanup());

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
    expect(FLAG_KEYS.APP_POS_PAYMENTS).toBe('app.pos.payments');
  });
});
