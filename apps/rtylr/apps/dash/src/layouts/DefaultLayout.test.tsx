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

/**
 * The anchors and the tour are written in two files, and nothing connected them
 * until the dashboard step silently attached to nothing for however long it took
 * to notice. This ties them together: every element a tour step attaches to must
 * be one this layout can actually render.
 */
describe('onboarding tour anchors', () => {
  it('every attachTo selector has a producer in the layout', async () => {
    const { buildTourSteps } = await import('../shared/hooks/useOnboardingTour');
    const { rawNavItems } = await import('./DefaultLayout');

    // The two anchors written directly into the layout's markup.
    const STATIC_ANCHORS = ['settings', 'caret'];
    const producible = new Set([...rawNavItems.map((i) => i.key), ...STATIC_ANCHORS]);

    const required = buildTourSteps(((k: string) => k) as never)
      .map((step) => {
        const el = (step.attachTo as { element?: string } | undefined)?.element;
        return el?.match(/\[data-tour="([^"]+)"\]/)?.[1];
      })
      .filter((id): id is string => Boolean(id));

    expect(required.length).toBeGreaterThan(0);
    expect(required.filter((id) => !producible.has(id))).toEqual([]);
  });
});
