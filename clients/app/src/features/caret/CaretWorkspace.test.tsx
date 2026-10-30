import path from 'node:path';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import type { ReactNode } from 'react';

const state = vi.hoisted(() => ({ owner: true, admin: false, mobile: false }));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: unknown) => unknown) =>
    selector({
      user: {
        id: 1,
        businessId: 2,
        effectiveAccess: { isOwner: state.owner, isAdmin: state.admin, apps: [] },
      },
    }),
}));
vi.mock('@voxire/caret', () => ({
  CaretIcon: () => null,
  useCaretStore: Object.assign(
    (selector: (state: unknown) => unknown) => selector({ tabs: [], activeTabId: null }),
    { getState: () => ({ newTab: vi.fn() }) }
  ),
  useCaretUIStore: (selector: (state: unknown) => unknown) =>
    selector({ view: 'chat', setView: vi.fn() }),
  CaretPanel: ({
    workspace,
  }: {
    workspace: {
      sidebarOpen: boolean;
      sidebarHeader: ReactNode;
      sidebarContent: ReactNode;
      mainContent: ReactNode;
      header: ReactNode;
      onSelectChat: () => void;
    };
  }) => (
    <div>
      {workspace.sidebarOpen && (
        <aside>
          {workspace.sidebarHeader}
          {workspace.sidebarContent}
          <button onClick={workspace.onSelectChat}>New chat</button>
        </aside>
      )}
      <main>{workspace.mainContent || workspace.header}</main>
    </div>
  ),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/lib/portalCaretContext', () => ({
  buildPortalPageContext: () => ({}),
  buildPortalSystemContext: () => '',
}));
vi.mock('../../shell/PortalHeaderActions', () => ({ usePortalHeaderActions: () => null }));
vi.mock('./useCaretInstall', () => ({
  useCaretInstall: () => ({ canInstall: false, install: vi.fn() }),
}));
vi.mock('./Buddy', () => ({
  default: ({ sidebarToggle }: { sidebarToggle: ReactNode }) => (
    <div>{sidebarToggle}Buddy conversation</div>
  ),
}));
import CaretWorkspace from './CaretWorkspace';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));
beforeEach(() => {
  state.owner = true;
  state.admin = false;
  state.mobile = false;
  window.matchMedia = vi.fn().mockImplementation(() => ({
    get matches() {
      return state.mobile;
    },
  }));
});
afterEach(cleanup);
describe('Caret Buddy sidebar', () => {
  it('opens Buddy in the workspace and returns to a new chat', async () => {
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    await user.click(screen.getByRole('button', { name: 'Caret Buddy' }));
    expect(screen.getByText('Buddy conversation')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Caret Buddy' })).toHaveAttribute(
      'aria-current',
      'page'
    );
    await user.click(screen.getByRole('button', { name: 'New chat' }));
    expect(screen.queryByText('Buddy conversation')).not.toBeInTheDocument();
  });
  it('hides Buddy from non-admins', () => {
    state.owner = false;
    render(<CaretWorkspace />);
    expect(screen.queryByRole('button', { name: 'Caret Buddy' })).not.toBeInTheDocument();
  });
  it('closes the mobile sidebar on selection and lets an admin reopen it', async () => {
    state.owner = false;
    state.admin = true;
    state.mobile = true;
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    await user.click(screen.getByRole('button', { name: 'Toggle chat sidebar' }));
    await user.click(screen.getByRole('button', { name: 'Caret Buddy' }));
    expect(screen.getByText('Buddy conversation')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Caret Buddy' })).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Toggle chat sidebar' }));
    expect(screen.getByRole('button', { name: 'Caret Buddy' })).toBeInTheDocument();
  });
});
