import path from 'node:path';
import { useState, type ReactNode } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../test/i18nTestInstance';
import { settingsNavItems } from './navigation';
import SettingsNavigation from './SettingsNavigation';
import { useUserStore } from '../shared/stores/userStore';
import { createEffectiveAccess } from '../../test/effectiveAccess';

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    to,
    search,
    children,
    ...props
  }: {
    to: string;
    search?: Record<string, string>;
    children: ReactNode;
  }) => (
    <a href={search ? `${to}?${new URLSearchParams(search)}` : to} {...props}>
      {children}
    </a>
  ),
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => unknown;
  }) => select({ location: { pathname: '/settings/flow' } }),
}));
const locales = path.resolve(import.meta.dirname, '../locales');
const flowLocales = [path.resolve(import.meta.dirname, '../modules/flow/locales')];
function Navigation() {
  const [query, setQuery] = useState('');
  return (
    <SettingsNavigation
      items={settingsNavItems}
      query={query}
      onQueryChange={setQuery}
      onNavigate={() => {}}
    />
  );
}
beforeEach(async () => {
  await initTestI18nFrom(locales, 'en', flowLocales);
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
});
describe('Settings navigation search', () => {
  it('filters app settings and restores grouped links when cleared', async () => {
    const ui = userEvent.setup();
    render(<Navigation />);
    expect(screen.queryByRole('link', { name: 'Settings', exact: true })).toBeNull();
    await ui.type(screen.getByRole('searchbox'), 'WORK');
    expect(screen.getByRole('link', { name: 'Work', exact: true })).toHaveAttribute(
      'href',
      '/settings/flow'
    );
    expect(screen.queryByRole('link', { name: 'Account', exact: true })).toBeNull();
    await ui.clear(screen.getByRole('searchbox'));
    expect(screen.getByRole('link', { name: 'Account', exact: true })).toBeVisible();
    expect(screen.getByRole('link', { name: 'Insights', exact: true })).toBeVisible();
  });
  it('places Flow panels in the shared sidebar and gates management links', () => {
    useUserStore.setState({ user: { effectiveAccess: createEffectiveAccess('flow') } });
    render(<Navigation />);
    expect(screen.getByRole('link', { name: 'Notifications', exact: true })).toHaveAttribute(
      'href',
      '/settings/flow?section=notifications'
    );
    expect(screen.getByRole('link', { name: 'Time', exact: true })).toHaveAttribute(
      'href',
      '/settings/flow?section=time'
    );
    expect(screen.queryByRole('link', { name: 'Team', exact: true })).toBeNull();
    expect(screen.queryByRole('link', { name: 'Recycle bin', exact: true })).toBeNull();
  });
  it('supports searching by settings section', async () => {
    const ui = userEvent.setup();
    render(<Navigation />);
    await ui.type(screen.getByRole('searchbox'), 'business');
    expect(screen.getByRole('link', { name: 'Billing', exact: true })).toBeVisible();
    expect(screen.queryByRole('link', { name: 'Work', exact: true })).toBeNull();
  });
  it.each([
    ['en', 'No settings found.'],
    ['fr', 'Aucun paramètre trouvé.'],
    ['ar', 'لم يتم العثور على إعدادات.'],
  ])('shows translated empty results in %s', async (language, message) => {
    await initTestI18nFrom(locales, language, flowLocales);
    const ui = userEvent.setup();
    render(<Navigation />);
    await ui.type(screen.getByRole('searchbox'), 'not-a-setting');
    expect(screen.getByRole('status')).toHaveTextContent(message);
  });
});
