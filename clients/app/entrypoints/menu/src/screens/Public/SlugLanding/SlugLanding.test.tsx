import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll, beforeEach } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

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

const businessQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: true,
}));
vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicBusinessBySlug: () => businessQuery,
}));

vi.mock('../Menu/Menu', () => ({
  default: () => null,
}));

import SlugLanding from './SlugLanding';
import { initTestI18n } from '../../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('SlugLanding business-not-found state', () => {
  beforeEach(() => {
    backSpy.mockClear();
    businessQuery.data = undefined;
    businessQuery.isLoading = false;
    businessQuery.isError = true;
  });

  it('shows the shared EmptyState and goes back on the CTA', () => {
    render(<SlugLanding />);

    expect(screen.getByText('Business not found')).toBeInTheDocument();
    expect(
      screen.getByText(/The link "\/no-such-business" doesn't match any active business/)
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));

    expect(backSpy).toHaveBeenCalledTimes(1);
  });
});

describe('SlugLanding location picker', () => {
  beforeEach(() => {
    businessQuery.isError = false;
    businessQuery.isLoading = false;
  });

  it('prompts for a location and lists the active ones', () => {
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [
        { id: 5, name: 'Downtown', slug: 'downtown', address: '1 Main St', timezone: 'UTC' },
        { id: 6, name: 'Harbour', slug: 'harbour', address: '', timezone: 'UTC' },
      ],
    };

    render(<SlugLanding />);

    expect(screen.getByText('Pick a location to see its menu')).toBeInTheDocument();
    expect(screen.getByText('Downtown')).toBeInTheDocument();
    expect(screen.getByText('Harbour')).toBeInTheDocument();
  });

  it('says so when a business has no active location', () => {
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [],
    };

    render(<SlugLanding />);

    expect(screen.getByText('No active locations yet.')).toBeInTheDocument();
  });
});
