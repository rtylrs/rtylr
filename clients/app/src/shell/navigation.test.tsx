import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Home } from 'lucide-react';
import { PortalNavLink } from './navigation';
const location = vi.hoisted(() => ({ pathname: '/team/groups/new' }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  useRouterState: ({ select }: { select: (state: { location: { pathname: string } }) => string }) =>
    select({ location }),
}));
describe('Portal navigation', () => {
  it('selects Groups exclusively for a nested group page and keeps translated tour anchors stable', () => {
    render(
      <>
        <PortalNavLink key="team" path="/team" label="Équipe" icon={Home} tourId="team" />
        <PortalNavLink key="groups" path="/team/groups" label="Groupes" icon={Home} />
      </>
    );
    expect(screen.getByRole('link', { name: 'Groupes' })).toHaveAttribute('aria-current', 'page');
    expect(screen.getByRole('link', { name: 'Équipe' })).not.toHaveAttribute('aria-current');
    expect(screen.getByRole('link', { name: 'Équipe' })).toHaveAttribute('data-tour', 'team');
  });
});
