import path from 'node:path';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
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
const store = vi.hoisted(() => ({ sessionId: null as string | null, openChat: vi.fn() }));
vi.mock('@voxire/caret', () => ({
  CaretIcon: () => null,
  useCaretStore: Object.assign(
    (selector: (state: unknown) => unknown) =>
      selector({
        tabs: [
          { id: 't', sessionId: store.sessionId, title: 'Chat', messages: [], isStreaming: false },
        ],
        activeTabId: 't',
      }),
    { getState: () => ({ newTab: vi.fn(), openChat: store.openChat }) }
  ),
  useCaretUIStore: (selector: (state: unknown) => unknown) =>
    selector({ view: 'chat', setView: vi.fn() }),
  CaretPanel: ({
    systemContext,
    workspace,
  }: {
    systemContext: string;
    workspace: {
      sidebarOpen: boolean;
      sidebarHeader: ReactNode;
      sidebarContent: ReactNode;
      header: ReactNode;
      pollSessionId?: string;
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
      <main>
        {workspace.header}
        <p data-testid="context">{systemContext}</p>
        <p data-testid="poll">{workspace.pollSessionId ?? ''}</p>
      </main>
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
const buddy = vi.hoisted(() => ({ getBuddyThread: vi.fn() }));
vi.mock('./BuddyPerch', () => ({ default: () => null }));
vi.mock('./Buddy', () => ({
  buddySystemContext: () => 'You are Caret Buddy',
  buddyClient: buddy,
  default: () => <div>Routines drawer</div>,
}));
import CaretWorkspaceView from './CaretWorkspace';

const CaretWorkspace = () => (
  <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
    <CaretWorkspaceView />
  </QueryClientProvider>
);

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
  beforeEach(() => {
    store.sessionId = null;
    store.openChat.mockReset();
    buddy.getBuddyThread.mockResolvedValue({ sessionId: 'buddy-2-1' });
  });

  it('opens the Buddy chat for any member', async () => {
    state.owner = false;
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    await user.click(screen.getByRole('button', { name: 'Caret Buddy' }));
    expect(store.openChat).toHaveBeenCalledWith({ sessionId: 'buddy-2-1', title: 'Caret Buddy' });
  });

  it('chats as Buddy, keeps the chat fresh and opens routines', async () => {
    store.sessionId = 'buddy-2-1';
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    expect(screen.getByRole('button', { name: 'Caret Buddy' })).toHaveAttribute(
      'aria-current',
      'page'
    );
    expect(screen.getByTestId('context')).toHaveTextContent('You are Caret Buddy');
    expect(screen.getByTestId('poll')).toHaveTextContent('buddy-2-1');
    await user.click(screen.getByRole('button', { name: 'Routines' }));
    expect(screen.getByText('Routines drawer')).toBeInTheDocument();
  });

  it('is a normal chat elsewhere', () => {
    store.sessionId = 'abc';
    render(<CaretWorkspace />);
    expect(screen.getByTestId('context')).not.toHaveTextContent('You are Caret Buddy');
    expect(screen.queryByRole('button', { name: 'Routines' })).not.toBeInTheDocument();
  });

  it('says so when the chat cannot open', async () => {
    buddy.getBuddyThread.mockRejectedValue(new Error('down'));
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    await user.click(screen.getByRole('button', { name: 'Caret Buddy' }));
    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't open Caret Buddy");
  });

  it('flags a Buddy post you have not seen', async () => {
    localStorage.clear();
    store.sessionId = 'abc';
    buddy.getBuddyThread.mockResolvedValue({
      sessionId: 'buddy-2-1',
      lastPostAt: new Date().toISOString(),
    });
    render(<CaretWorkspace />);
    expect(
      await screen.findByRole('status', { name: 'New message from Caret Buddy' })
    ).toBeInTheDocument();
  });

  it('clears the flag once the Buddy chat is open', async () => {
    localStorage.clear();
    store.sessionId = 'buddy-2-1';
    buddy.getBuddyThread.mockResolvedValue({
      sessionId: 'buddy-2-1',
      lastPostAt: new Date(Date.now() - 1000).toISOString(),
    });
    render(<CaretWorkspace />);
    await screen.findByRole('button', { name: 'Routines' });
    expect(screen.queryByRole('status', { name: 'New message from Caret Buddy' })).toBeNull();
  });

  it('closes the mobile sidebar when Buddy opens', async () => {
    state.mobile = true;
    const user = userEvent.setup();
    render(<CaretWorkspace />);
    await user.click(screen.getByRole('button', { name: 'Toggle chat sidebar' }));
    await user.click(screen.getByRole('button', { name: 'Caret Buddy' }));
    expect(screen.queryByRole('button', { name: 'Caret Buddy' })).not.toBeInTheDocument();
  });
});
