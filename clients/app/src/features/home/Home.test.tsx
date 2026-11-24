import React from 'react';
import path from 'node:path';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../test/effectiveAccess';
import Home from './Home';
import { RENNILI_URL, resolveAppUrl } from '@voxire/config';
import { useUserStore } from '../../shared/stores/userStore';
import { useCommandPalette } from '../../shared/stores/commandPaletteStore';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
}));

describe('Portal Home', () => {
  beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));
  beforeEach(() => {
    useUserStore.setState({ user: null });
    useCommandPalette.setState({ open: false });
  });
  it.each([{ isOwner: true }, { isAdmin: true }])(
    'offers team and groups to a manager %j',
    (role) => {
      useUserStore.setState({
        user: {
          id: 1,
          fullName: 'Taylor Test',
          businessName: 'Example',
          effectiveAccess: createEffectiveAccess('flow', role),
        },
      });
      render(<Home />);
      expect(screen.getByRole('heading', { name: 'Welcome back, Taylor' })).toBeVisible();
      expect(screen.getByRole('link', { name: /Members and access/ })).toHaveAttribute(
        'href',
        '/team'
      );
      expect(screen.getByRole('link', { name: /Shared permissions/ })).toHaveAttribute(
        'href',
        '/team/groups'
      );
      fireEvent.click(screen.getByRole('button', { name: /Search apps/ }));
      expect(useCommandPalette.getState().open).toBe(true);
    }
  );
  it('keeps a member in Home, shows only authorized shortcuts and never automatically launches Flow', () => {
    useUserStore.setState({
      user: { id: 2, fullName: 'Member Test', effectiveAccess: createEffectiveAccess('flow') },
    });
    render(<Home />);
    expect(screen.getByRole('heading', { name: 'Welcome back, Member' })).toBeVisible();
    expect(screen.queryByRole('link', { name: /Members and access/ })).toBeNull();
    expect(screen.queryByRole('link', { name: /Shared permissions/ })).toBeNull();
    expect(screen.getByRole('link', { name: /Your profile and preferences/ })).toHaveAttribute(
      'href',
      '/settings/account'
    );
    expect(screen.getByRole('link', { name: 'Work', exact: true })).toHaveAttribute(
      'href',
      '/flow'
    );
    expect(screen.getByRole('link', { name: 'Work', exact: true })).not.toHaveAttribute('target');
  });
  it('launches POS at its own origin while keeping other apps in Portal', () => {
    useUserStore.setState({ user: { effectiveAccess: createEffectiveAccess('pos') } });
    render(<Home />);
    expect(screen.getByRole('link', { name: /Open Point of Sale/ })).toHaveAttribute(
      'href',
      resolveAppUrl('pos')
    );
  });
  it('opens Rennili board in a new tab for every signed-in business member', () => {
    useUserStore.setState({
      user: {
        id: 2,
        role: 'owner',
        effectiveAccess: createEffectiveAccess('flow', { licensed: false }),
      } as never,
    });
    render(<Home />);
    expect(screen.queryByRole('link', { name: /Members and access/ })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Work', exact: true })).toBeNull();
    const rennili = screen.getByRole('link', { name: /Open Rennili/ });
    expect(rennili).toHaveAttribute('href', new URL('/board', RENNILI_URL).href);
    expect(rennili).toHaveAttribute('target', '_blank');
    expect(rennili).toHaveAttribute('rel', 'noopener noreferrer');
  });
  it('shows the no-apps message when no user is loaded', () => {
    render(<Home />);
    expect(screen.queryByRole('link', { name: /Open Rennili/ })).toBeNull();
    expect(screen.getByText(/No apps are assigned/)).toBeVisible();
  });
});
