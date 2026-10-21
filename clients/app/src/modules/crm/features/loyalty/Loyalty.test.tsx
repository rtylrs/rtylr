import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const configState = vi.hoisted(() => ({
  data: undefined as unknown,
  isError: false,
}));

const updateConfig = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const auth = vi.hoisted(() => ({ effectiveAccess: null as unknown }));

vi.mock('../../../../shared/stores/userStore', () => {
  const state = () => ({
    user: { effectiveAccess: auth.effectiveAccess },
    effectiveAccess: auth.effectiveAccess,
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useLoyaltyConfig: () => ({ data: configState.data, isError: configState.isError }),
  useLoyaltyTransactions: () => ({ data: [] }),
  useLoyaltyTiers: () => ({ data: [] }),
  useUpdateLoyaltyConfig: () => updateConfig,
}));

import Loyalty from './Loyalty';

const loadedConfig = {
  programEnabled: true,
  pointsPerDollar: 1,
  pointsExpiry: 365,
  redemptionRate: 0.01,
};

const saveButton = () => screen.getByRole('button', { name: 'Save' });

beforeEach(() => {
  auth.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
  configState.data = undefined;
  configState.isError = false;
  updateConfig.mutate.mockClear();
});

describe('Loyalty config load gating (B60-R06)', () => {
  it('disables Save and the program toggle until the config has loaded', () => {
    render(<Loyalty />);

    expect(saveButton()).toBeDisabled();
    expect(screen.getByRole('button', { name: /Program Paused/ })).toBeDisabled();
  });

  it('surfaces a config load failure instead of silently showing defaults', () => {
    configState.isError = true;
    render(<Loyalty />);

    expect(screen.getByRole('alert')).toHaveTextContent('crm.loyalty.configLoadFailed');
    expect(saveButton()).toBeDisabled();
  });

  it('enables Save once the config has loaded', () => {
    configState.data = loadedConfig;
    render(<Loyalty />);

    expect(saveButton()).toBeEnabled();
  });
});
