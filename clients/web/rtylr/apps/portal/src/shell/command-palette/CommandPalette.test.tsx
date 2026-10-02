import path from 'node:path';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useCommandPalette } from '../../shared/stores/commandPaletteStore';
import { useUserStore } from '../../shared/stores/userStore';
import { useKeyboardShortcuts } from '../../shared/hooks/useKeyboardShortcuts';
import CommandPalette from './CommandPalette';

const navigate = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigate,
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => unknown;
  }) => select({ location: { pathname: '/' } }),
}));
vi.mock('../../modules/flow/useFlowCommands', () => ({
  useFlowCommands: () => ({ commands: [], loading: false, failed: false }),
}));

const locales = path.resolve(import.meta.dirname, '../../locales');

function Launcher() {
  useKeyboardShortcuts();
  return (
    <>
      <button onClick={() => useCommandPalette.getState().setOpen(true)}>Search Portal</button>
      <a href="#outside">Outside link</a>
      <CommandPalette />
    </>
  );
}

beforeEach(async () => {
  vi.restoreAllMocks();
  navigate.mockReset();
  await initTestI18nFrom(locales);
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
  useCommandPalette.setState({ open: false });
});

describe('Portal command launcher', () => {
  it.each([{ isOwner: true }, { isAdmin: true }])(
    'unlocks subscribed apps and management actions for %j without per-page grants',
    async (role) => {
      const ui = userEvent.setup();
      useUserStore.setState({
        user: {
          effectiveAccess: createEffectiveAccess('erp', {
            ...role,
            licensed: false,
            hasAccess: false,
          }),
        },
      });
      render(<Launcher />);
      await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
      expect(screen.getByRole('option', { name: 'ERP' })).toBeVisible();
      expect(screen.getByRole('option', { name: 'Groups' })).toBeVisible();
      expect(screen.getByRole('option', { name: 'Add a member' })).toBeVisible();
    }
  );
  it('opens with Cmd+K, contains focus, and restores it when dismissed', async () => {
    const ui = userEvent.setup();
    render(<Launcher />);
    const trigger = screen.getByRole('button', { name: 'Search Portal' });
    trigger.focus();
    await ui.keyboard('{Meta>}k{/Meta}');
    const input = screen.getByRole('combobox', { name: 'Search apps, pages and actions…' });
    expect(input).toHaveFocus();
    expect(screen.getByRole('dialog', { name: 'Search Portal' })).toBeVisible();
    await ui.tab();
    expect(screen.getByRole('button', { name: 'Close dialog' })).toHaveFocus();
    await ui.tab();
    expect(input).toHaveFocus();
    await ui.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(trigger).toHaveFocus();
  });

  it('executes the hovered page with Enter, including items before the last result', async () => {
    const ui = userEvent.setup();
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    const groups = screen.getByRole('option', { name: 'Groups' });
    fireEvent.pointerMove(groups);
    expect(groups).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-activedescendant', groups.id);
    await ui.keyboard('{Enter}');
    expect(navigate).toHaveBeenCalledWith({ to: '/team/groups' });
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('filters permissions while open and removes actions when access changes', async () => {
    const ui = userEvent.setup();
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('flow', { dashPermissions: ['hub:team:view'] }),
      },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    expect(screen.getByRole('option', { name: 'Flow' })).toBeVisible();
    expect(screen.getByRole('option', { name: 'Team' })).toBeVisible();
    expect(screen.queryByRole('option', { name: 'Add a member' })).toBeNull();
    expect(screen.queryByRole('option', { name: 'ERP' })).toBeNull();
    expect(screen.queryByRole('option', { name: 'Insights' })).toBeNull();
    expect(screen.queryByRole('option', { name: 'Menus' })).toBeNull();
    act(() =>
      useUserStore.setState({
        user: { effectiveAccess: createEffectiveAccess('flow', { licensed: false }) },
      })
    );
    expect(screen.queryByRole('option', { name: 'Flow' })).toBeNull();
    expect(screen.queryByRole('option', { name: 'Team' })).toBeNull();
  });

  it('offers an invite to a member with invite permission without granting team read access', async () => {
    const ui = userEvent.setup();
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess(undefined, { dashPermissions: ['hub:team:invite'] }),
      },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    expect(screen.queryByRole('option', { name: 'Team' })).toBeNull();
    await ui.type(screen.getByRole('combobox'), 'invite');
    expect(screen.getByRole('option', { name: 'Add a member' })).toBeVisible();
    await ui.keyboard('{Enter}');
    expect(navigate).toHaveBeenCalledWith({ to: '/team/invite' });
  });

  it('opens Insights inside Portal at the first page the member can use', async () => {
    const ui = userEvent.setup();
    const external = vi.spyOn(window, 'open').mockImplementation(() => null);
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('insights', {
          permissions: ['insights:reports:create'],
        }),
      },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    expect(screen.queryByRole('option', { name: 'Overview' })).toBeNull();
    expect(screen.getByRole('option', { name: 'Reports' })).toBeVisible();
    await ui.click(screen.getByRole('option', { name: 'Insights' }));
    expect(navigate).toHaveBeenCalledWith({ to: '/insights/reports' });
    expect(external).not.toHaveBeenCalled();
  });

  it('offers delegated audit and developer access while hiding unrelated management pages', async () => {
    const ui = userEvent.setup();
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('insights', {
          permissions: ['insights:dashboard:view'],
          dashPermissions: ['hub:audit:view', 'hub:api_keys:manage'],
        }),
      },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    expect(screen.getByRole('option', { name: 'Audit Log' })).toBeVisible();
    expect(screen.getByRole('option', { name: 'API Keys' })).toBeVisible();
    expect(screen.getByRole('option', { name: 'Webhooks' })).toBeVisible();
    expect(screen.queryByRole('option', { name: 'Add a member' })).toBeNull();
    await ui.click(screen.getByRole('option', { name: 'Audit Log' }));
    expect(navigate).toHaveBeenCalledWith({ to: '/audit' });
  });

  it('searches app descriptions and opens migrated ERP inside Portal', async () => {
    const ui = userEvent.setup();
    useUserStore.setState({
      user: { effectiveAccess: createEffectiveAccess('erp', { isOwner: true }) },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    await ui.type(screen.getByRole('combobox'), 'inventory');
    await ui.click(screen.getByRole('option', { name: /ERP/ }));
    expect(navigate).toHaveBeenCalledWith({ to: '/erp' });
  });

  it('safely opens an app that has not moved into Portal', async () => {
    const ui = userEvent.setup();
    const external = vi.spyOn(window, 'open').mockImplementation(() => null);
    useUserStore.setState({
      user: { effectiveAccess: createEffectiveAccess('finance', { isOwner: true }) },
    });
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    await ui.click(screen.getByRole('option', { name: /^Finance/ }));
    expect(external).toHaveBeenCalledWith(
      expect.stringContaining('5182'),
      '_blank',
      'noopener,noreferrer'
    );
    expect(navigate).not.toHaveBeenCalled();
  });

  it('offers a single Extensions command in Workspace', async () => {
    const ui = userEvent.setup();
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    await ui.type(screen.getByRole('combobox'), 'extensions');
    expect(screen.getAllByRole('option')).toHaveLength(1);
    await ui.click(screen.getByRole('option', { name: 'Extensions' }));
    expect(navigate).toHaveBeenCalledWith({ to: '/extensions' });
  });

  it('uses canonical account and billing destinations without duplicate audit entries', async () => {
    const ui = userEvent.setup();
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    await ui.type(screen.getByRole('combobox'), 'audit');
    expect(screen.getAllByRole('option')).toHaveLength(1);
    await ui.click(screen.getByRole('button', { name: 'Clear search' }));
    await ui.click(screen.getByRole('option', { name: 'Account' }));
    expect(navigate).toHaveBeenLastCalledWith({ to: '/settings/account' });
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    await ui.type(screen.getByRole('combobox'), 'billing');
    await ui.keyboard('{Enter}');
    expect(navigate).toHaveBeenLastCalledWith({ to: '/settings/billing' });
  });

  it.each([
    ['en', 'No results for “nothing-matches”', 'Clear search'],
    ['fr', 'Aucun résultat pour « nothing-matches »', 'Effacer la recherche'],
    ['ar', 'لا توجد نتائج لـ «nothing-matches»', 'مسح البحث'],
  ])('renders the empty state and clears the query in %s', async (lang, message, clearLabel) => {
    await initTestI18nFrom(locales, lang);
    const ui = userEvent.setup();
    render(<Launcher />);
    await ui.click(screen.getByRole('button', { name: 'Search Portal' }));
    const input = screen.getByRole('combobox');
    await ui.type(input, 'nothing-matches');
    expect(screen.getByRole('status')).toHaveTextContent(message);
    await ui.keyboard('{ArrowDown}{Enter}');
    expect(navigate).not.toHaveBeenCalled();
    await ui.click(screen.getAllByRole('button', { name: clearLabel })[0]);
    expect(input).toHaveValue('');
    expect(input).toHaveFocus();
    expect(screen.getAllByRole('option').length).toBeGreaterThan(0);
  });
});
