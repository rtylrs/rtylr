import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const uploadAssetMock = vi.hoisted(() => vi.fn());

vi.mock('@voxire/api', () => ({
  uploadAsset: uploadAssetMock,
}));

vi.mock('@voxire/auth', () => ({
  setCookie: vi.fn(),
  removeCookie: vi.fn(),
  getDefaultCookieDomain: () => 'rtylr.com',
}));

type Business = { id: number; name: string; isOwned?: boolean; isActive?: boolean };

const businessesData = vi.hoisted(() => ({
  data: { businesses: [] as Business[] } as { businesses: Business[] } | undefined,
  isLoading: false,
}));

const setBusinessLogoMock = vi.hoisted(() => ({ mutateAsync: vi.fn(), isPending: false }));

vi.mock('../../../../../shared/hooks/queries/useBusinesses', () => ({
  useListMyBusinesses: () => businessesData,
  useCreateSubBusiness: () => ({ mutate: vi.fn(), isPending: false }),
  useRenameBusiness: () => ({ mutate: vi.fn(), isPending: false }),
  useSetBusinessLogo: () => setBusinessLogoMock,
  useDeleteBusiness: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useSwitchBusiness: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { businessId: 1, businessName: 'Acme' } }),
}));

vi.mock('../../../../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => ({ data: undefined }),
  useBillingDetail: () => ({ data: undefined }),
}));

import Businesses from './Businesses';

beforeEach(() => {
  vi.clearAllMocks();
  businessesData.data = { businesses: [] };
  businessesData.isLoading = false;
});

describe('Businesses list empty state', () => {
  it('shows the illustration, title and an add-business call-to-action when there are no businesses', () => {
    render(<Businesses />);

    expect(screen.getByText('businesses.noBusinesses')).toBeInTheDocument();
  });

  it('opens the create-business dialog when the call-to-action is clicked', () => {
    render(<Businesses />);

    expect(screen.queryByText('businesses.createDescription')).not.toBeInTheDocument();

    const addButtons = screen.getAllByRole('button', { name: 'businesses.addBusiness' });
    fireEvent.click(addButtons[addButtons.length - 1]);

    expect(screen.getByText('businesses.createDescription')).toBeInTheDocument();
  });

  it('does not show the empty state when there are businesses', () => {
    businessesData.data = { businesses: [{ id: 1, name: 'Acme' }] };
    render(<Businesses />);

    expect(screen.queryByText('businesses.noBusinesses')).not.toBeInTheDocument();
    // "Acme" is the only business and also the current business, so it renders in
    // both the list row and the current-business side panel.
    expect(screen.getAllByText('Acme').length).toBeGreaterThan(0);
  });
});

describe('Businesses logo upload', () => {
  it('uploads the logo picked for a second business while the first upload is still in flight', async () => {
    businessesData.data = {
      businesses: [
        { id: 1, name: 'Acme' },
        { id: 2, name: 'Beta' },
      ],
    };

    let finishFirstUpload!: (value: { url: string }) => void;
    uploadAssetMock
      .mockImplementationOnce(
        () =>
          new Promise<{ url: string }>((resolve) => {
            finishFirstUpload = resolve;
          })
      )
      .mockImplementationOnce(() => Promise.resolve({ url: 'url-b' }));
    setBusinessLogoMock.mutateAsync.mockResolvedValue({});

    const { container } = render(<Businesses />);
    const fileInput = container.querySelector('input[type="file"]') as HTMLInputElement;
    const logoButtons = screen.getAllByRole('button', { name: /changeLogo/ });

    fireEvent.click(logoButtons[0]);
    Object.defineProperty(fileInput, 'files', {
      value: [new File(['a'], 'a.png', { type: 'image/png' })],
      configurable: true,
    });
    fireEvent.change(fileInput);

    // Open the second business's picker while the first upload is unresolved.
    fireEvent.click(logoButtons[1]);

    await act(async () => {
      finishFirstUpload({ url: 'url-a' });
    });
    await waitFor(() =>
      expect(setBusinessLogoMock.mutateAsync).toHaveBeenCalledWith({
        businessId: 1,
        logoUrl: 'url-a',
      })
    );

    Object.defineProperty(fileInput, 'files', {
      value: [new File(['b'], 'b.png', { type: 'image/png' })],
      configurable: true,
    });
    fireEvent.change(fileInput);

    await waitFor(() =>
      expect(setBusinessLogoMock.mutateAsync).toHaveBeenCalledWith({
        businessId: 2,
        logoUrl: 'url-b',
      })
    );
    expect(uploadAssetMock).toHaveBeenCalledTimes(2);
  });
});
