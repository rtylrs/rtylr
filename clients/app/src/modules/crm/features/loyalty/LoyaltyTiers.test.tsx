import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const hookState = vi.hoisted(() => ({
  tiers: [] as unknown[],
}));

const effectiveAccess = () => ({
  isOwner: true,
  subscribedApps: ['crm'],
  apps: [{ appId: 'crm', licensed: true, hasAccess: true }],
});

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('../../../../shared/stores/userStore', () => {
  const state = () => ({
    user: { effectiveAccess: effectiveAccess() },
    effectiveAccess: effectiveAccess(),
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useLoyaltyTiers: () => ({ data: hookState.tiers, isLoading: false }),
  useUpsertTier: () => ({ mutate: vi.fn(), isPending: false }),
}));

import LoyaltyTiers from './LoyaltyTiers';

const tier = (id: number, threshold: number) => ({
  id,
  name: `Tier ${threshold}`,
  threshold,
  multiplier: 1,
  color: '#000000',
  perks: [],
});

describe('LoyaltyTiers ordering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the tiers in ascending threshold order', () => {
    hookState.tiers = [tier(1, 500), tier(2, 100), tier(3, 250)];

    render(<LoyaltyTiers />);

    const names = screen
      .getAllByRole('heading')
      .map((el) => el.textContent)
      .filter((name): name is string => typeof name === 'string');
    // The tier headings only; the page header carries its own heading.
    expect(names.filter((name) => name.startsWith('Tier '))).toEqual([
      'Tier 100',
      'Tier 250',
      'Tier 500',
    ]);
  });

  // `Array#sort` sorts in place, and the array here is the one held in the
  // react-query cache, so sorting it directly rewrote the cached order for
  // every other consumer of the same query key.
  it('leaves the cached array it was given untouched', () => {
    const cached = [tier(1, 500), tier(2, 100), tier(3, 250)];
    hookState.tiers = cached;

    render(<LoyaltyTiers />);

    expect(cached.map((t) => t.threshold)).toEqual([500, 100, 250]);
  });
});
