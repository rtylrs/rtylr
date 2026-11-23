import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import { RTYLR_CONTACT_CHANNELS } from '@voxire/config';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    to,
    children,
    ...rest
  }: {
    to: string;
    children: React.ReactNode;
    [key: string]: unknown;
  }) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
}));

vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

import Landing from './Landing';
import FeaturesPage from './FeaturesPage';
import RtylrServicesPage from './RtylrServicesPage';
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Rennili landing page', () => {
  it('routes primary CTAs to /try and venue setup to auth registration', () => {
    render(<Landing />);

    const demoLinks = screen.getAllByRole('link', { name: /Try the live demo/i });
    expect(demoLinks.length).toBeGreaterThanOrEqual(2);
    for (const link of demoLinks) {
      expect(link).toHaveAttribute('href', '/try');
    }

    const setupLinks = screen.getAllByRole('link', { name: 'Set up my venue' });
    expect(setupLinks.length).toBeGreaterThanOrEqual(1);
    for (const link of setupLinks) {
      const url = new URL(link.getAttribute('href') ?? '');
      expect(url.pathname).toBe('/register');
      expect(url.searchParams.get('returnTo')).toBe('/pager/setup');
    }

    const signInUrl = new URL(
      screen.getByRole('link', { name: 'Sign in' }).getAttribute('href') ?? ''
    );
    expect(signInUrl.pathname).toBe('/');
    expect(signInUrl.searchParams.get('returnTo')).toBe('/pager/setup');
  });

  it('opens the mobile navigation sheet, traps focus, and closes on Escape', () => {
    render(<Landing />);

    const menuButton = screen.getByRole('button', { name: 'Open menu' });
    expect(menuButton).toHaveAttribute('aria-expanded', 'false');

    fireEvent.click(menuButton);
    const dialog = screen.getByRole('dialog', { name: 'Main navigation' });
    expect(dialog).toBeInTheDocument();
    expect(menuButton).toHaveAttribute('aria-expanded', 'true');

    const closeButton = within(dialog).getByRole('button', { name: 'Close menu' });
    expect(closeButton).toHaveFocus();

    // Shift+Tab from the first focusable item wraps to the last inside the dialog.
    fireEvent.keyDown(document, { key: 'Tab', shiftKey: true });
    const dialogLinks = within(dialog).getAllByRole('link');
    expect(dialogLinks[dialogLinks.length - 1]).toHaveFocus();

    // Escape closes the sheet and restores focus to the trigger button.
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(menuButton).toHaveFocus();
  });

  it('rings the hero simulator and toggles FAQ disclosures', () => {
    render(<Landing />);

    const tapReady = screen.getByRole('button', { name: 'Tap Ready' });
    expect(tapReady).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(tapReady);
    expect(screen.getByRole('button', { name: 'Reset preview' })).toHaveAttribute(
      'aria-pressed',
      'true'
    );

    const secondQuestionButton = screen.getByRole('button', {
      name: 'Does it work with my existing POS, or without one?',
    });
    expect(secondQuestionButton).toHaveAttribute('aria-expanded', 'false');
    fireEvent.click(secondQuestionButton);
    expect(secondQuestionButton).toHaveAttribute('aria-expanded', 'true');
  });

  it('sets each plastic buzzer problem beside what replaces it', () => {
    render(<Landing />);

    const table = screen.getByRole('table');
    const [head, ...rows] = within(table).getAllByRole('row');
    expect(
      within(head)
        .getAllByRole('columnheader')
        .map((cell) => cell.textContent)
    ).toEqual(['Plastic pager rack', 'Rennili']);
    expect(rows).toHaveLength(5);
    expect(
      within(rows[0])
        .getAllByRole('cell')
        .map((cell) => cell.textContent)
    ).toEqual([
      'Buy a full set upfront and keep replacing broken units',
      'Nothing to buy: the guest brings the device',
    ]);
  });

  it('keeps the match code shuffler and the mini board on the features page', async () => {
    render(<FeaturesPage />);

    fireEvent.click(screen.getByRole('button', { name: 'Next code' }));
    expect(await screen.findByText('318')).toBeInTheDocument();

    const markReady = screen.getByRole('button', { name: 'Mark ready' });
    expect(markReady).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(markReady);
    expect(screen.getByRole('button', { name: 'Reset' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByText('Coffee shops & cafés')).toBeInTheDocument();
  });

  it('links navigation items to dedicated SEO pages and shows the RTYLR services page', () => {
    render(<Landing />);

    const nav = screen.getByRole('navigation', { name: 'Main navigation' });
    expect(within(nav).getByRole('link', { name: 'How it works' })).toHaveAttribute(
      'href',
      '/how-it-works'
    );
    expect(within(nav).getByRole('link', { name: 'Features' })).toHaveAttribute(
      'href',
      '/features'
    );
    expect(within(nav).getByRole('link', { name: 'By RTYLR' })).toHaveAttribute('href', '/rtylr');
    expect(within(nav).getByRole('link', { name: 'FAQ' })).toHaveAttribute('href', '/faq');

    expect(screen.getByRole('link', { name: /Explore RTYLR services/ })).toHaveAttribute(
      'href',
      '/rtylr'
    );

    // Dedicated RTYLR services upsell page
    const { unmount } = render(<RtylrServicesPage />);
    expect(screen.getByText('One connected platform for your entire venue')).toBeInTheDocument();
    expect(screen.getAllByText('Point of Sale (POS)').length).toBeGreaterThanOrEqual(1);
    unmount();
  });

  it('renders the shared RTYLR contact channels in the footer', () => {
    render(<Landing />);

    const footer = screen.getByRole('contentinfo');
    for (const channel of RTYLR_CONTACT_CHANNELS) {
      const links = within(footer).getAllByRole('link');
      expect(links.some((link) => link.getAttribute('href') === channel.href)).toBe(true);
    }
  });
});
