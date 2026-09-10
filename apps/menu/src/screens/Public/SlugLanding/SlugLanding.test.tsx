import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeAll } from 'vitest';

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

vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicBusinessBySlug: () => ({ data: undefined, isLoading: false, isError: true }),
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
