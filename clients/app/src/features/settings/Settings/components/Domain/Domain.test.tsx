import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const toastError = vi.hoisted(() => vi.fn());
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: toastError },
}));

const customDomainState = vi.hoisted(() => ({
  data: undefined as
    { domain?: string; status?: string; verificationToken?: string } | null | undefined,
}));
const registerState = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
const verifyState = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
const removeState = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));

vi.mock('../../../../../shared/hooks/queries/useBusiness', () => ({
  useCustomDomain: () => customDomainState,
  useRegisterCustomDomain: () => registerState,
  useVerifyCustomDomain: () => verifyState,
  useRemoveCustomDomain: () => removeState,
}));

vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { businessName: 'Acme' } }),
}));

vi.mock('../../../../../shared/stores/configStore', () => ({
  useConfigStore: () => ({ slug: 'acme' }),
}));

import Domain from './Domain';

beforeEach(() => {
  vi.clearAllMocks();
  customDomainState.data = undefined;
});

describe('Domain: registered domain', () => {
  it('seeds the input and a pending badge from the custom-domain query', () => {
    customDomainState.data = {
      domain: 'menu.acme.com',
      status: 'pending_verification',
      verificationToken: 'challenge-token',
    };
    render(<Domain />);

    expect(screen.getByDisplayValue('menu.acme.com')).toBeInTheDocument();
    expect(screen.getByText('settings.domain.statusPending')).toBeInTheDocument();
    expect(screen.queryByText('settings.domain.statusVerified')).not.toBeInTheDocument();
  });

  it('shows the TXT challenge to publish and verifies with its token', () => {
    customDomainState.data = {
      domain: 'menu.acme.com',
      status: 'pending_verification',
      verificationToken: 'challenge-token',
    };
    render(<Domain />);

    expect(screen.getByText('_rtylr-challenge.menu.acme.com')).toBeInTheDocument();
    expect(screen.getByText('challenge-token')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'settings.domain.verifyButton' }));
    expect(verifyState.mutate).toHaveBeenCalledWith('challenge-token');
  });

  it('shows a verified domain without the verify step', () => {
    customDomainState.data = { domain: 'menu.acme.com', status: 'verified' };
    render(<Domain />);

    expect(screen.getByText('settings.domain.statusVerified')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'settings.domain.verifyButton' })
    ).not.toBeInTheDocument();
  });

  it('removes the claim', () => {
    customDomainState.data = { domain: 'menu.acme.com', status: 'verified' };
    render(<Domain />);

    fireEvent.click(screen.getByRole('button', { name: 'settings.domain.removeButton' }));
    expect(removeState.mutate).toHaveBeenCalled();
  });
});

describe('Domain: connect', () => {
  it('registers the typed domain through the custom-domain mutation', () => {
    render(<Domain />);

    fireEvent.change(screen.getByLabelText('settings.domain.domainLabel'), {
      target: { value: ' Menu.Acme.COM ' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'settings.domain.connectDomain' }));

    // Pre-fix this called updateBusiness with {customDomain, website}; the
    // hook dropped customDomain and the page toasted a success it did not have.
    expect(registerState.mutate).toHaveBeenCalledWith('menu.acme.com');
  });

  it('rejects an invalid domain without calling the API', () => {
    render(<Domain />);

    fireEvent.change(screen.getByLabelText('settings.domain.domainLabel'), {
      target: { value: 'not a domain' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'settings.domain.connectDomain' }));

    expect(registerState.mutate).not.toHaveBeenCalled();
    expect(toastError).toHaveBeenCalled();
  });
});
