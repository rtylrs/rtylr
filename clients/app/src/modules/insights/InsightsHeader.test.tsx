import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const routerState = vi.hoisted(() => ({ pathname: '/insights' }));

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    to,
    ...rest
  }: { children: React.ReactNode; to: string } & React.HTMLAttributes<HTMLAnchorElement>) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
  useRouterState: ({ select }: { select: (s: { location: { pathname: string } }) => string }) =>
    select({ location: { pathname: routerState.pathname } }),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/ui', () => ({
  DropdownMenu: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DropdownMenuTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DropdownMenuContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DropdownMenuItem: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (select: (s: { user: null }) => unknown) => select({ user: null }),
}));

vi.mock('../../shared/utils/canAccessPortalPath', () => ({
  canAccessPortalPath: () => true,
}));

vi.mock('../../shared/components/RtylrLogo', () => ({ default: () => null }));

import InsightsHeader from './InsightsHeader';

function brand(): string {
  return screen.getByRole('link', { name: 'appNames.short.insights' }).textContent ?? '';
}

describe('InsightsHeader', () => {
  beforeEach(() => {
    routerState.pathname = '/insights';
  });

  it('lists Projects among the Insights pages', () => {
    render(<InsightsHeader actions={null} />);
    expect(screen.getAllByText('insightsNav.projects').length).toBeGreaterThan(0);
  });

  it('names a project dashboard page after Projects, not Overview', () => {
    routerState.pathname = '/insights/projects/4';
    render(<InsightsHeader actions={null} />);
    expect(brand()).toContain('insightsNav.projects');
  });

  it('keeps Overview for the Insights root', () => {
    render(<InsightsHeader actions={null} />);
    expect(brand()).toContain('insightsNav.overview');
  });
});
