import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  effectiveAccess: null as unknown,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => {
  const state = () => ({
    user: { effectiveAccess: mocks.effectiveAccess },
    effectiveAccess: mocks.effectiveAccess,
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

const refetchRewards = vi.fn();
const rewardsList: unknown[] = [];

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useLoyaltyRewards: () => ({
    data: rewardsList,
    isLoading: false,
    refetch: refetchRewards,
  }),
  useUpsertReward: () => ({ mutate: vi.fn(), isPending: false }),
}));

import LoyaltyRewards from './LoyaltyRewards';

function renderRewards(): void {
  render(<LoyaltyRewards />);
}

beforeEach(() => {
  refetchRewards.mockClear();
  mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: [] });
});

describe('LoyaltyRewards empty state', () => {
  it('shows an "add reward" call-to-action for a role that can configure loyalty', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
    renderRewards();

    expect(screen.getByText('crm.rewards.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.rewards.emptyDescription')).toBeInTheDocument();

    // Header action and empty-state action are the same action offered twice and
    // now share an accessible name; assert both exist rather than matching one.
    const addRewardButtons = screen.getAllByRole('button', { name: 'crm.rewards.addReward' });
    expect(addRewardButtons).toHaveLength(2);
    fireEvent.click(addRewardButtons[1]);

    // Opens the inline "New reward" form rather than mutating anything itself.
    expect(screen.getByText('crm.rewards.newReward')).toBeInTheDocument();
  });

  it('falls back to a "refresh" call-to-action for a role that cannot configure loyalty', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: [] });
    renderRewards();

    expect(screen.getByText('crm.rewards.empty')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchRewards).toHaveBeenCalledTimes(1);
  });
});
