import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  clearConfig: vi.fn(),
  signOut: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  signOut: mocks.signOut,
}));

vi.mock('../stores/hrConfigStore', () => ({
  useHRConfigStore: {
    getState: () => ({ clearConfig: mocks.clearConfig }),
  },
}));

import { signOutFromHR } from './hrSignOut';

describe('signOutFromHR', () => {
  beforeEach(() => {
    mocks.clearConfig.mockReset();
    mocks.signOut.mockReset();
  });

  it('clears persisted HR config before redirecting through shared auth', () => {
    signOutFromHR();

    expect(mocks.clearConfig).toHaveBeenCalledOnce();
    expect(mocks.signOut).toHaveBeenCalledWith('hr');
    expect(mocks.clearConfig.mock.invocationCallOrder[0]).toBeLessThan(
      mocks.signOut.mock.invocationCallOrder[0]
    );
  });
});
