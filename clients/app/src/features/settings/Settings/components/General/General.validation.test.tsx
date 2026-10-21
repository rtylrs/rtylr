import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
  AnimatePresence: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('../../../../../shared/stores/businessStore', () => ({
  useBusinessStore: () => ({ businessType: 'restaurant' }),
}));

const updateBusinessMock = vi.hoisted(() => ({ mutateAsync: vi.fn(), isPending: false }));
const setSlugMock = vi.hoisted(() => ({ mutateAsync: vi.fn(), isPending: false }));

vi.mock('../../../../../shared/hooks/queries/useBusiness', () => ({
  useUpdateBusiness: () => updateBusinessMock,
  useSetBusinessSlug: () => setSlugMock,
}));

const userState = vi.hoisted(() => ({
  user: { email: 'acme@example.com', businessName: 'Acme' } as Record<string, unknown>,
}));

vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: userState.user }),
}));

const configState = vi.hoisted(() => ({
  websiteData: { slug: 'acme' } as Record<string, unknown> | undefined,
  contactInfo: undefined as Record<string, unknown> | undefined,
}));

vi.mock('../../../../../shared/stores/configStore', () => ({
  useConfigStore: () => ({
    websiteData: configState.websiteData,
    contactInfo: configState.contactInfo,
  }),
}));

import General from './General';

beforeEach(() => {
  vi.clearAllMocks();
  configState.websiteData = { slug: 'acme' };
});

const slugInput = (): HTMLInputElement =>
  screen.getByPlaceholderText('settings.general.slugPlaceholder') as HTMLInputElement;
const saveButton = (): HTMLElement => screen.getByRole('button', { name: 'common.saveChanges' });

describe('General: public menu URL presentation', () => {
  it('shows the Portal origin and the /menu suffix instead of the retired menu host', () => {
    render(<General />);

    expect(screen.getByText('/menu')).toBeInTheDocument();
    expect(screen.getByText(/localhost:5175\//)).toBeInTheDocument();
    expect(screen.queryByText(/menu\.rtylr\.com/)).not.toBeInTheDocument();
  });
});

describe('General: reserved slug', () => {
  it('shows an inline error and blocks saving a reserved slug', () => {
    render(<General />);

    fireEvent.change(slugInput(), { target: { value: 'menu' } });

    expect(screen.getByRole('alert')).toHaveTextContent('settings.general.slugReserved');
    expect(saveButton()).toBeDisabled();

    fireEvent.click(saveButton());
    expect(setSlugMock.mutateAsync).not.toHaveBeenCalled();
  });

  it('saves a non-reserved slug without an error', async () => {
    render(<General />);

    fireEvent.change(slugInput(), { target: { value: 'acme-2' } });

    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(saveButton()).toBeEnabled();

    fireEvent.click(saveButton());
    await waitFor(() => expect(setSlugMock.mutateAsync).toHaveBeenCalledWith('acme-2'));
  });
});
