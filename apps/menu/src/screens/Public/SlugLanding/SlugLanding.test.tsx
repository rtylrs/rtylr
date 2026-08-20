import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key}:${JSON.stringify(opts)}` : key,
  }),
}));

vi.mock('@voxire/components', () => ({
  FullScreenLoader: () => null,
  HelmetContainer: () => null,
}));

const backSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'no-such-business' }),
  useNavigate: () => vi.fn(),
  useRouter: () => ({ history: { back: backSpy } }),
  Link: ({ children }: { children?: React.ReactNode }) => <a>{children}</a>,
}));

vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicBusinessBySlug: () => ({ data: undefined, isLoading: false, isError: true }),
}));

vi.mock('../Menu/Menu', () => ({
  default: () => null,
}));

import SlugLanding from './SlugLanding';

describe('SlugLanding business-not-found state', () => {
  it('shows the shared EmptyState and goes back on the CTA', () => {
    render(<SlugLanding />);

    expect(screen.getByText('businessNotFoundTitle')).toBeInTheDocument();
    expect(screen.getByText(/businessNotFoundDescription/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'goBack' }));

    expect(backSpy).toHaveBeenCalledTimes(1);
  });
});
