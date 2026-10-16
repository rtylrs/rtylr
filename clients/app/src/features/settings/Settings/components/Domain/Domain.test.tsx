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

const customDomainState = vi.hoisted(() => ({ data: undefined as string | undefined }));
const registerState = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));

vi.mock('../../../../../shared/hooks/queries/useBusiness', () => ({
  useCustomDomain: () => customDomainState,
  useRegisterCustomDomain: () => registerState,
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
  it('seeds the input and badge from the custom-domain query', () => {
    customDomainState.data = 'menu.acme.com';
    render(<Domain />);

    expect(screen.getByDisplayValue('menu.acme.com')).toBeInTheDocument();
    expect(screen.getByText('settings.domain.configured')).toBeInTheDocument();
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
