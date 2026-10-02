import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...rest }: { to: string; children: React.ReactNode }) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
}));

import RouteNotFound from './RouteNotFound';

describe('RouteNotFound', () => {
  it('is a designed screen rather than raw router text', () => {
    // Without a defaultNotFoundComponent, TanStack renders the words "Not
    // Found" at the top left of a white page, outside the app frame and
    // untranslated.
    render(<RouteNotFound />);

    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('flow.notFound.title');
    expect(screen.getByText('flow.notFound.description')).toBeInTheDocument();
  });

  it('offers exactly one action, and it leads somewhere real', () => {
    render(<RouteNotFound />);

    const links = screen.getAllByRole('link');
    expect(links).toHaveLength(1);
    expect(links[0]).toHaveAttribute('href', '/flow/my-issues');
    expect(links[0]).toHaveTextContent('flow.notFound.cta');
  });

  it('goes through i18n, so it is not untranslated English', () => {
    render(<RouteNotFound />);
    expect(screen.queryByText('Not Found')).not.toBeInTheDocument();
    // Every visible string is a key here, which means every one went through t().
    for (const key of ['flow.notFound.title', 'flow.notFound.description', 'flow.notFound.cta']) {
      expect(screen.getByText(key)).toBeInTheDocument();
    }
  });

  it('keeps the sentence readable at a wide width', () => {
    const { container } = render(<RouteNotFound />);
    const sentence = container.querySelector('p');
    expect(sentence?.className).toContain('max-w-[44ch]');
  });
});
