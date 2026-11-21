import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  useStaff: vi.fn(),
  effectiveAccess: undefined as unknown,
}));

vi.mock('./useStaff', () => ({
  useStaff: mocks.useStaff,
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      selector: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      selector({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  Link: ({ children }: { children: React.ReactNode }) => <a>{children}</a>,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import Staff from './Staff';

describe('Staff page directory restriction handling', () => {
  beforeEach(() => {
    mocks.useStaff.mockReset();
    mocks.effectiveAccess = createEffectiveAccess('hr', {
      permissions: ['hr:employees:view', 'hr:time_off:request'],
    });
  });

  it('shows a clear restricted message instead of 0 team members when directory access is 403', () => {
    mocks.useStaff.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      isFetching: false,
      refetch: vi.fn(),
      error: Object.assign(new Error('The company directory is turned off for this business'), {
        status: 403,
      }),
    });

    render(<Staff />);

    expect(screen.getByText('hrStaff.restricted.title')).toBeInTheDocument();
    expect(screen.getByText('hrStaff.restricted.description')).toBeInTheDocument();
    expect(screen.queryByText('hrStaff.memberCount')).not.toBeInTheDocument();
    expect(screen.queryByText('hrStaff.inviteStaff')).not.toBeInTheDocument();
  });
});
