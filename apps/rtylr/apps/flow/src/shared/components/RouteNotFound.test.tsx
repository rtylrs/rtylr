import fs from 'node:fs';
import path from 'node:path';
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
    expect(links[0]).toHaveAttribute('href', '/my-issues');
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

describe('RouteNotFound is actually wired into the router', () => {
  // A mutation probe caught this gap: deleting defaultNotFoundComponent from
  // createRouter left every test above green, because they cover the screen
  // and not the wiring. Without it the screen is unreachable and TanStack's
  // raw "Not Found" renders instead, so the whole fix is inert.
  const router = fs.readFileSync(path.resolve(__dirname, '../../router.tsx'), 'utf8');

  it('registers the screen as the router default', () => {
    expect(router).toContain('defaultNotFoundComponent');
    expect(router).toMatch(/defaultNotFoundComponent:\s*\(\)\s*=>\s*withLayout\(<RouteNotFound/);
  });

  it('renders it inside the app frame, not bare', () => {
    // Bare, the 404 loses the sidebar and navigation, which is the reported
    // defect rather than only the missing copy.
    const match = router.match(/defaultNotFoundComponent:\s*\(\)\s*=>\s*([^,\n]+)/);
    expect(match?.[1]).toContain('withLayout(');
  });

  it('imports the screen it names', () => {
    expect(router).toContain("import RouteNotFound from './shared/components/RouteNotFound'");
  });
});
