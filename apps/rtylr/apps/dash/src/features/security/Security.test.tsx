import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-qrcode-logo', () => ({
  QRCode: () => null,
}));

const securityState = vi.hoisted(() => ({
  data: undefined as Record<string, unknown> | undefined,
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useSecurity', () => ({
  useSecurity: () => securityState,
  useToggleTwoFactor: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useBeginMFAEnrollment: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useCompleteMFAEnrollment: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useVerifyMFA: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useMFARecoveryCode: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useRevokeSession: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: { user: null }) => unknown) => selector({ user: null }),
}));

import Security from './Security';

const loadedState = {
  mfaEnrolled: false,
  twoFactorEnabled: false,
  mfaRequired: false,
  sessions: [],
  trustedDevices: [],
};

describe('Security page load states', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    securityState.data = undefined;
    securityState.isLoading = false;
    securityState.isError = false;
  });

  it('renders an error panel with retry instead of the skeleton when the query fails', () => {
    securityState.isError = true;
    const { container } = render(<Security />);

    expect(screen.getByText('Failed to load security settings.')).toBeInTheDocument();
    expect(container.querySelectorAll('.animate-pulse')).toHaveLength(0);

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(securityState.refetch).toHaveBeenCalledTimes(1);
  });

  it('still renders the skeleton while the query is loading', () => {
    securityState.isLoading = true;
    const { container } = render(<Security />);

    expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0);
    expect(screen.queryByText('Failed to load security settings.')).not.toBeInTheDocument();
  });

  it('renders the page once the security state loads', () => {
    securityState.data = loadedState;
    render(<Security />);

    expect(screen.getByText('Active sessions')).toBeInTheDocument();
    expect(screen.getByText('Trusted devices')).toBeInTheDocument();
  });
});
