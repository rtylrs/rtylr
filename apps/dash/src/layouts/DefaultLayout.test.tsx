import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import React from 'react';

vi.mock('@tanstack/react-router', () => ({
  // `to` is the router's prop name; an <a> needs href to carry the link role.
  Link: ({ children, to, ...rest }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
  useRouterState: () => '/',
}));

import { SidebarNavItem } from './DefaultLayout';

/**
 * The onboarding tour attaches to `[data-tour="dashboard"]`, `"team"` and
 * `"apps"` (useOnboardingTour.ts). The anchor used to be `label.toLowerCase()`,
 * which is a translated string, so it was both wrong in English and about to
 * break in every other language:
 *
 *   nav.dashboard is "Account Home" -> data-tour="account home"
 *
 * which no tour step matches. Keying it off the stable nav key fixes that and
 * makes it immune to translation.
 */
describe('SidebarNavItem tour anchor', () => {
  const icon = (): React.JSX.Element => <span />;

  it('uses the stable key, not the label, so the dashboard step can find it', () => {
    render(
      <SidebarNavItem
        tourId="dashboard"
        // The English label the anchor used to be derived from.
        label="Account Home"
        path="/"
        key="dashboard"
        icon={icon}
      />
    );

    expect(screen.getByRole('link')).toHaveAttribute('data-tour', 'dashboard');
  });

  it('does not change when the label is translated', () => {
    render(<SidebarNavItem tourId="team" label="الفريق" path="/team" key="team" icon={icon} />);

    expect(screen.getByRole('link')).toHaveAttribute('data-tour', 'team');
  });
});
