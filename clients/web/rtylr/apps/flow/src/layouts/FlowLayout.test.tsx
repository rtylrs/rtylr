import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import FlowLayout from './FlowLayout';
import { useFlowStore } from '../shared/stores/flowStore';
import { createEffectiveAccess } from '../../../../test/effectiveAccess';

// FlowLayout mounts the whole app shell (sidebar, header, realtime wiring,
// caret panel, command bar, etc.). This test only exercises the account menu
// (avatar/name trigger in the sidebar footer - issue #785: it didn't close on
// Escape or on an outside click), so every react-query-backed hook/component
// FlowLayout pulls in is stubbed out. Store hooks (userStore, appearanceStore,
// flowStore) and other pure/local hooks (useKeyboardShortcuts,
// useLifePreferences, useForesightLinks) are left real - none of them touch the
// network and they're already exercised by other passing tests in this app.
const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  signOut: vi.fn(),
  location: { pathname: '/inbox' },
  // A stable object reference: FlowLayout mirrors this into its local
  // userStore via a `useEffect` keyed on `[authUser, setUser]`. If the mock
  // returned a fresh object on every render, that effect would think the
  // dependency changed every time and call setUser in a loop.
  // Effective access is updated in place so the auth-user object identity stays
  // stable for the effect described above.
  authUser: {
    id: '1',
    name: 'Test User',
    email: 'test@example.com',
    effectiveAccess: null as ReturnType<typeof createEffectiveAccess> | null,
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  useLocation: () => mocks.location,
  useNavigate: () => mocks.navigate,
  useParams: () => ({}),
  // The layout reads `searchStr`; the TaskDrawer it renders reads `search`.
  // Supplying only one of them made the drawer's selector throw on
  // `state.location.search.openTask`, which surfaced as every test in this file
  // failing rather than as anything about a drawer.
  useRouterState: ({
    select,
  }: {
    select: (state: {
      location: { searchStr: string; search: Record<string, unknown>; pathname: string };
    }) => unknown;
  }) => select({ location: { searchStr: '', search: {}, pathname: '/' } }),
}));

vi.mock('motion/react', () => {
  const Motion = ({
    children,
    className,
    style,
    onClick,
  }: React.PropsWithChildren<{
    className?: string;
    style?: React.CSSProperties;
    onClick?: React.MouseEventHandler;
  }>) => (
    <div className={className} style={style} onClick={onClick}>
      {children}
    </div>
  );

  return {
    AnimatePresence: ({ children }: React.PropsWithChildren) => <>{children}</>,
    motion: { div: Motion, main: Motion, span: Motion, button: Motion },
  };
});

vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
  getAvatarUrl: () => '',
  resolveAppUrl: () => '',
  // Pulled in transitively: '@voxire/api' resolves its base URL at module
  // scope when the generated clients are constructed.
  resolveApiBase: () => 'https://api.test/',
}));

const providerProps = vi.hoisted(
  () =>
    ({ workspaceAccent: undefined }) as {
      workspaceAccent: string | undefined;
    }
);

vi.mock('@voxire/ui', () => ({
  Avatar: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
  // The real provider writes custom properties to documentElement, which this
  // test has no opinion about: the token values belong to the package's own
  // suite. What does belong here is the one prop this layout is responsible for
  // supplying, so the stub records it rather than discarding it.
  AppearanceProvider: ({
    children,
    workspaceAccent,
  }: React.PropsWithChildren<{ workspaceAccent?: string }>) => {
    providerProps.workspaceAccent = workspaceAccent;
    return <>{children}</>;
  },
}));

// caretState.isOpen is mutable so a test can assert the panel is mounted only
// while open, and CaretPanel renders a textarea because the defect was Caret's
// composer holding keyboard focus while the panel was closed.
const caretState = vi.hoisted(() => ({ isOpen: false }));
vi.mock('@voxire/caret', () => ({
  buildCaretRouteContext: () => ({}),
  CaretPanel: () => <textarea aria-label="Ask a question about this page" />,
  CaretButton: ({ className }: { className?: string }) => (
    <button type="button" className={className}>
      Caret
    </button>
  ),
  useCaretStore: () => caretState,
}));

// useForesightLinks is stubbed because it prefetches routes; useClickOutside is
// deliberately the REAL hook, since the outside-click cases below are what it
// exists to prove. Stubbing it would make those tests assert nothing.
vi.mock('@voxire/utils', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@voxire/utils')>()),
  useForesightLinks: () => undefined,
}));

vi.mock('@voxire/auth', () => ({
  signOut: mocks.signOut,
  useUserStore: (
    select: (state: {
      user: typeof mocks.authUser;
      effectiveAccess: typeof mocks.authUser.effectiveAccess;
    }) => unknown
  ) => select({ user: mocks.authUser, effectiveAccess: mocks.authUser.effectiveAccess }),
}));

// Sidebar preferences use React Query in production. The account-menu and
// appearance tests do not exercise that persistence path, so keep the shell
// isolated from its provider and network requirements here.
vi.mock('../shared/lib/sidebarPreferences', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../shared/lib/sidebarPreferences')>();
  const preferences = actual.createDefaultSidebarPreferences();
  const update = vi.fn();

  return {
    ...actual,
    useSidebarPreferences: () => ({
      preferences,
      update,
      isLoading: false,
      isSaving: false,
    }),
  };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../shared/components/TrialBanner', () => ({ default: () => null }));
vi.mock('../shared/components/CommandBar', () => ({ default: () => null }));
vi.mock('../shared/components/QuickCreate', () => ({ default: () => null }));
vi.mock('../shared/components/BulkEditBar', () => ({ default: () => null }));
// The bin section runs its own react-query hook; this suite renders the layout
// without a QueryClientProvider, like the other data-fetching children above.

vi.mock('../features/workspace-dashboard/useWorkspaces', () => ({
  useWorkspaces: () => ({ data: undefined, isLoading: false }),
}));
const createProjectMutate = vi.hoisted(() => vi.fn());
vi.mock('../features/projects/useProjects', () => ({
  useProjects: () => ({ data: undefined }),
  useCreateProject: () => ({ mutate: createProjectMutate, isPending: false }),
}));
vi.mock('../features/tasks/useMyTasksProject', () => ({
  useMyTasksProject: () => ({ data: undefined }),
}));
vi.mock('../features/inbox/useNotifications', () => ({
  useNotifications: () => ({ data: [] }),
}));
vi.mock('../shared/hooks/useFlowBusinessName', () => ({
  useFlowBusinessName: () => ({ data: null }),
  useFlowBusinessId: () => ({ data: '42' }),
}));
// FlowLayout reads real contracted hours per member (#868). Stubbed like every
// other react-query hook here: unmocked it needs a QueryClientProvider, and an
// empty map is the "no capacity known" case these tests do not exercise.
vi.mock('../shared/hooks/queries/useMemberCapacity', () => ({
  useMemberCapacityHours: () => ({ data: {} }),
}));

vi.mock('../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [] }),
}));
vi.mock('../shared/lib/useLayoutRealtime', () => ({
  useLayoutRealtime: () => undefined,
}));
vi.mock('../shared/lib/realtime', () => ({
  useRealtimeReconnectRefetch: () => undefined,
  useRealtimeStatus: () => 'disconnected',
}));
vi.mock('../shared/lib/realtimeRouteGate', () => ({
  routeNeedsRealtime: () => false,
}));

describe('FlowLayout account menu', () => {
  beforeEach(() => {
    mocks.navigate.mockClear();
    mocks.signOut.mockClear();
    mocks.location = { pathname: '/inbox' };
  });

  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  it('opens on trigger click and shows the sign-out action', () => {
    renderLayout();

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));

    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();
  });

  it('closes on Escape', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();
  });

  it('closes when clicking outside the menu', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();

    fireEvent.mouseDown(document.body);

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();
  });

  it('stays open when clicking inside the menu', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    const signOutButton = screen.getByText('flow.shell.signOut');

    fireEvent.mouseDown(signOutButton);

    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();
  });
});

// The 'FlowLayout workspace switcher' suite is deliberately gone.
//
// This PR removes workspace state from the layout: activeWorkspace is now a
// literal built from the business name, and there is no switcher to open, no
// list to enumerate and nothing to close. The four tests here asserted a control
// that no longer renders, so they were removed with it rather than left to fail
// or skipped as if the feature were coming back.
//
// If a workspace switcher returns, the assertions are recoverable from git
// history at this commit's parent.

describe('FlowLayout appearance wiring', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  // A colour that is deliberately NOT ACCENT_DEFAULT, so the assertions below
  // can tell a live value from a constant. That mattered when the layout read a
  // real workspace: the fallback workspace's colour IS ACCENT_DEFAULT, so an
  // earlier version of this suite passed with the wiring replaced by a constant.
  const DISTINCT = '#a3d84b';

  beforeEach(() => {
    providerProps.workspaceAccent = undefined;
    useFlowStore.getState().setWorkspaces([]);
  });

  // This suite changed shape with the workspace removal, and the change is a
  // product decision worth reading rather than a test repair.
  //
  // The layout used to pass the active workspace's colour, which is what made
  // `accent: null` mean "follow the workspace" in the appearance panel. It now
  // builds activeWorkspace as a literal with `color: ACCENT_DEFAULT`, so the
  // accent the provider receives is constant.
  //
  // The consequence for the appearance feature: the user-accent-overrides-
  // workspace rule still holds, but its other half is gone. `accent: null` now
  // resolves to the brand teal rather than to anything workspace-derived. The
  // panel's reset control still works and the derivation is unaffected. Nothing
  // is broken; one half of a documented behaviour is simply no longer reachable.
  //
  // Asserting the constant rather than deleting the suite, so a future change
  // that reintroduces a real per-workspace accent has to come back here and say
  // so deliberately.
  it('hands the appearance provider a constant accent now that workspaces are gone', () => {
    useFlowStore.getState().setWorkspaces([{ id: 'ws-9', name: 'Studio', color: DISTINCT }]);
    useFlowStore.getState().setActiveWorkspace('ws-9');
    renderLayout();
    expect(providerProps.workspaceAccent).not.toBe(DISTINCT);
    expect(providerProps.workspaceAccent).toBeTruthy();
  });

  it('still supplies the provider an accent at all', () => {
    renderLayout();
    expect(providerProps.workspaceAccent).toBeTruthy();
  });
});

describe('FlowLayout Caret panel mounting (#1043)', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  afterEach(() => {
    caretState.isOpen = false;
  });

  it('does not mount the composer while the panel is closed', () => {
    caretState.isOpen = false;

    renderLayout();

    // The panel used to render at width 0 with overflow hidden, so this
    // textarea existed, autofocused, and swallowed every single-key shortcut.
    expect(screen.queryByRole('textbox', { name: /Ask a question/i })).not.toBeInTheDocument();
  });

  it('leaves nothing focusable from the closed panel', () => {
    caretState.isOpen = false;

    const { container } = renderLayout();

    expect(container.querySelector('textarea')).toBeNull();
  });

  it('mounts the composer once the panel is open', () => {
    caretState.isOpen = true;

    renderLayout();

    expect(screen.getByRole('textbox', { name: /Ask a question/i })).toBeInTheDocument();
  });

  it('keeps the Caret trigger available whether the panel is open or closed', () => {
    caretState.isOpen = false;

    renderLayout();

    expect(screen.getAllByRole('button', { name: 'Caret' }).length).toBeGreaterThan(0);
  });
});

describe('FlowLayout sidebar wiring (#1212)', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  /**
   * A direct assertion that the sidebar the user actually sees is mounted.
   *
   * FlowLayout carried a second, inline sidebar behind a literal `false` for
   * long enough that a fix was landed against it with passing unit tests and no
   * runtime effect at all (#1076). Deleting that block left this one as the only
   * sidebar, and until now nothing said so out loud: the coverage was incidental,
   * via the new-project tests below, which reach the create control through it.
   * A test that fails for the right reason is worth more than four that fail for
   * a reason nobody stated.
   */
  it('mounts the sidebar navigation', () => {
    renderLayout();
    expect(screen.getByRole('button', { name: 'flow.shell.newProject' })).toBeTruthy();
  });
});

describe('FlowLayout New project modal (#1060)', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  const openModal = (): void => {
    fireEvent.click(screen.getByRole('button', { name: 'flow.shell.newProject' }));
  };

  const fill = (name: string, description: string): void => {
    fireEvent.change(screen.getByPlaceholderText('flow.shell.projectNamePlaceholder'), {
      target: { value: name },
    });
    fireEvent.change(screen.getByPlaceholderText('flow.shell.projectDescriptionPlaceholder'), {
      target: { value: description },
    });
  };

  beforeEach(() => {
    createProjectMutate.mockClear();
    mocks.authUser.effectiveAccess = createEffectiveAccess('flow', { isOwner: true });
  });

  afterEach(() => {
    mocks.authUser.effectiveAccess = null;
  });

  it('sends the description and the chosen visibility', () => {
    renderLayout();
    openModal();
    fill('QC Delta Two', 'DESC SHOULD PERSIST');

    // The Public option in the visibility radiogroup.
    const [, publicOption] = screen.getAllByRole('radio');
    fireEvent.click(publicOption);

    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    expect(createProjectMutate).toHaveBeenCalledTimes(1);
    expect(createProjectMutate.mock.calls[0][0]).toMatchObject({
      name: 'QC Delta Two',
      description: 'DESC SHOULD PERSIST',
      visibility: 'public',
    });
  });

  it('defaults to private when the user does not choose', () => {
    renderLayout();
    openModal();
    fill('Quiet project', 'Some notes');

    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    expect(createProjectMutate.mock.calls[0][0]).toMatchObject({ visibility: 'private' });
  });

  it('sends no description rather than an empty string when the field is blank', () => {
    renderLayout();
    openModal();
    fill('No description', '   ');

    fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    // The contract marks description @required and its validator rejects '',
    // so a blank field has to travel as undefined for the API layer to
    // substitute its sentinel.
    expect(createProjectMutate.mock.calls[0][0]).toMatchObject({ description: undefined });
  });

  it('sends the same payload from the Enter key path', () => {
    renderLayout();
    openModal();
    fill('Via Enter', 'ENTER PATH DESC');

    fireEvent.keyDown(screen.getByPlaceholderText('flow.shell.projectNamePlaceholder'), {
      key: 'Enter',
    });

    expect(createProjectMutate).toHaveBeenCalledTimes(1);
    expect(createProjectMutate.mock.calls[0][0]).toMatchObject({
      name: 'Via Enter',
      description: 'ENTER PATH DESC',
      visibility: 'private',
    });
  });
});

describe('FlowLayout desktop create control (#1076)', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  it('offers exactly two new-task buttons, one per breakpoint', () => {
    renderLayout();
    // Both are in the DOM; CSS decides which is visible, and happy-dom applies
    // none, so the classes are what the assertion below actually checks.
    const buttons = screen.getAllByRole('button', { name: 'flow.shell.newTask' });
    expect(buttons).toHaveLength(2);
  });

  it('shows one of them at desktop, where the header + is display:none', () => {
    renderLayout();
    const buttons = screen.getAllByRole('button', { name: 'flow.shell.newTask' });

    // The header + lives in an lg:hidden container, so at 1440px it measured
    // 0x0 and the desktop shell had no create control at all. The sidebar one
    // is lg:flex, so exactly one of the pair is visible at any width.
    const desktopOnly = buttons.filter((b) => b.className.includes('lg:flex'));
    expect(desktopOnly).toHaveLength(1);
    expect(desktopOnly[0].className).toContain('hidden');
  });

  it('opens quick create when the desktop control is clicked', () => {
    renderLayout();
    const desktop = screen
      .getAllByRole('button', { name: 'flow.shell.newTask' })
      .find((b) => b.className.includes('lg:flex'));
    expect(desktop).toBeDefined();

    expect(useFlowStore.getState().quickCreateOpen).toBe(false);
    fireEvent.click(desktop as HTMLElement);
    expect(useFlowStore.getState().quickCreateOpen).toBe(true);
    useFlowStore.getState().setQuickCreateOpen(false);
  });

  it('does not hide the create control behind a hover', () => {
    // The + is meant to be the most visible control on the panel, unlike the
    // ... menus, which reveal on hover.
    renderLayout();
    const desktop = screen
      .getAllByRole('button', { name: 'flow.shell.newTask' })
      .find((b) => b.className.includes('lg:flex'));
    expect(desktop?.className).not.toContain('flow-reveal-on-hover');
  });
});
