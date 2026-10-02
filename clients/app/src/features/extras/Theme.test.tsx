import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const uploadAssetMock = vi.hoisted(() => vi.fn());
vi.mock('@voxire/api', () => ({
  uploadAsset: uploadAssetMock,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

const toastError = vi.hoisted(() => vi.fn());
vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: toastError },
}));

const updateBusinessState = vi.hoisted(() => ({
  mutate: vi.fn(),
  isPending: false,
}));
vi.mock('../../shared/hooks/queries/useBusiness', () => ({
  useUpdateBusiness: () => updateBusinessState,
}));

const configState = vi.hoisted(() => ({
  websiteData: { siteName: 'Acme', logoUrl: '' } as Record<string, unknown>,
}));
vi.mock('../../shared/stores/configStore', () => ({
  useConfigStore: (selector: (state: { websiteData: Record<string, unknown> }) => unknown) =>
    selector({ websiteData: configState.websiteData }),
}));

import ThemeCustomizer from './Theme';

beforeEach(() => {
  vi.clearAllMocks();
  configState.websiteData = { siteName: 'Acme', logoUrl: '' };
});

describe('Theme: save', () => {
  it('sends the edited name as storeName and drops the unsupported accent', () => {
    render(<ThemeCustomizer />);

    fireEvent.change(screen.getByDisplayValue('Acme'), { target: { value: 'Acme Bakery' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save Theme' }));

    // Pre-fix the payload carried primaryColor and brandName, both of which
    // useUpdateBusiness silently dropped before the request went out.
    expect(updateBusinessState.mutate).toHaveBeenCalledWith({
      data: { storeName: 'Acme Bakery', logoUrl: '' },
    });
  });

  it('does not show an accent control that nothing persists or applies', () => {
    render(<ThemeCustomizer />);

    expect(screen.queryByText('Accent color')).not.toBeInTheDocument();
  });
});

describe('Theme: logo upload', () => {
  it('reports a failed upload to the user', async () => {
    uploadAssetMock.mockRejectedValue(new Error('upload failed'));
    const { container } = render(<ThemeCustomizer />);

    const input = container.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(['x'], 'logo.png', { type: 'image/png' });
    Object.defineProperty(input, 'files', { value: [file], configurable: true });
    fireEvent.change(input);

    await waitFor(() => expect(toastError).toHaveBeenCalledWith('Logo upload failed'));
  });
});
