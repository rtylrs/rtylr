import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';

// FlowSettings statically imports TimeSection, whose own import chain reaches
// the app's real i18next bootstrap module (shared/services/i18next.ts), which
// calls `.use(initReactI18next)` at import time - so this mock must keep the
// real exports and only override useTranslation, unlike the simpler
// object-literal mock other Flow tests use.
vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({ t: (key: string) => key }),
  };
});

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@voxire/config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/config')>();
  return {
    ...actual,
    useFeatureFlag: () => false,
    FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
  };
});

// Same reason as the react-i18next mock above: TimeSection's import chain
// (via WorkTypesSection -> useWorkDay) calls createOneKitClient at module
// load time, so the real exports must survive alongside the unwrap override.
vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return {
    ...actual,
    unwrap: () => Promise.resolve({}),
  };
});

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getUserSettings: () => Promise.resolve({}),
    updateUserSettings: () => Promise.resolve({}),
  },
}));

const flowTeamData = vi.hoisted(() => ({
  data: [] as Array<{
    id: string;
    name: string;
    email: string;
    role?: string;
    apps?: string[];
    isActive?: boolean;
  }>,
}));
const grantMutate = vi.hoisted(() => vi.fn());
const deactivateMutate = vi.hoisted(() => vi.fn());
vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: flowTeamData.data, isLoading: false }),
  useFlowGrantAccess: () => ({ mutate: grantMutate, isPending: false }),
  useFlowDeactivateMember: () => ({ mutate: deactivateMutate, isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowSubscription', () => ({
  useFlowSubscription: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  computeSeatProration: () => ({
    daysRemaining: 0,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'January',
  }),
}));

vi.mock('../../features/workspace-dashboard/useWorkspaces', () => ({
  useWorkspaces: () => ({ data: [], isLoading: false }),
  useCreateWorkspace: () => ({ mutate: vi.fn(), isPending: false }),
}));

const labelsData = vi.hoisted(() => ({
  data: [] as Array<{ id: number; name: string; color: string }>,
}));
const createLabelMutate = vi.hoisted(() => vi.fn());
vi.mock('../../features/tasks/useLabels', () => ({
  useLabels: () => ({ data: labelsData.data, isLoading: false }),
  useCreateLabel: () => ({ mutate: createLabelMutate, isPending: false }),
  useDeleteLabel: () => ({ mutate: vi.fn(), isPending: false }),
}));

// A mutable "cookie" the tests flip per role, decoded by FlowSettings' own
// getRoleFromToken exactly as it would decode a real auth cookie.
const tokenState = vi.hoisted(() => ({ token: null as string | null }));
const authState = vi.hoisted(() => ({
  effectiveAccess: undefined as EffectiveAccess | undefined,
  // The legacy business role, which is all getMe sends when it does not send
  // effectiveAccess. Kept separate so a test can present one without the other.
  user: undefined as { role?: string } | undefined,
}));
vi.mock('@voxire/auth', () => ({
  getCookie: () => tokenState.token,
  useUserStore: (selector: (state: typeof authState) => unknown): unknown => selector(authState),
}));

// A mutable stand-in for `?section=...`. FlowSettings reads this via
// settingsRouteApi.useSearch() - this is the exact hole #793 flagged: the
// value is user-controlled and must be validated against the role gate.
const searchState = vi.hoisted(() => ({ section: undefined as string | undefined }));
vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => searchState }),
}));

import FlowSettings, { TeamSection, LabelsSection } from './FlowSettings';

function makeToken(role: string): string {
  const payload = btoa(JSON.stringify({ role }));
  return `header.${payload}.sig`;
}

function makeFlowAccess(permissions: EffectiveAccess['dashPermissions']): EffectiveAccess {
  return {
    businessId: '10',
    dashRole: 'admin',
    dashRoleId: null,
    dashPermissions: [],
    subscribedApps: ['flow'],
    licenses: [
      {
        appId: 'flow',
        status: 'active',
        source: 'explicit',
        groupId: null,
        grantedAt: '2026-01-01T00:00:00Z',
        revokedAt: null,
      },
    ],
    apps: [
      {
        appId: 'flow',
        licensed: true,
        hasAccess: true,
        roleIds: [],
        primaryRoleId: null,
        displayRole: 'admin',
        permissions,
        scopes: [],
        sources: ['explicit'],
        permissionSources: {},
      },
    ],
    isOwner: false,
    isAdmin: true,
    authorizationVersion: 2,
  };
}

function renderSettings() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <FlowSettings />
    </QueryClientProvider>
  );
}

function renderTeamSection() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <TeamSection />
    </QueryClientProvider>
  );
}

function renderLabelsSection() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <LabelsSection />
    </QueryClientProvider>
  );
}

// #793 finding 2 & 3: TeamSection was listed unconditionally in the sidebar
// (any role could grant/revoke Flow access), and `?section=workspaces` /
// `?section=time` rendered role-gated sections for any role because
// activeSection was never checked against the same gate that hides the tab.
describe('FlowSettings section gating (#793)', () => {
  beforeEach(() => {
    flowTeamData.data = [];
    grantMutate.mockClear();
    deactivateMutate.mockClear();
    searchState.section = undefined;
    authState.effectiveAccess = undefined;
    authState.user = undefined;
  });

  it('shows Team, Workspaces and Time in the sidebar for an admin', () => {
    tokenState.token = makeToken('admin');
    authState.effectiveAccess = makeFlowAccess(['flow:settings:manage', 'flow:team:manage']);
    renderSettings();

    expect(screen.getByRole('button', { name: 'common.settings.team' })).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'flow.settings.workspaces.section' })
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.time.section' })).toBeInTheDocument();
  });

  it('hides Team, Workspaces and Time from the sidebar for a plain member', () => {
    tokenState.token = makeToken('member');
    renderSettings();

    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.workspaces.section' })
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.time.section' })
    ).not.toBeInTheDocument();
  });

  it('hides Team from an unrecognized business-wide role (operator/staff/etc.)', () => {
    tokenState.token = makeToken('store_manager');
    renderSettings();

    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
  });

  it('does not render WorkspacesSection for a member who reaches it via ?section=workspaces', () => {
    tokenState.token = makeToken('member');
    searchState.section = 'workspaces';
    renderSettings();

    expect(
      screen.queryByRole('heading', { name: 'flow.settings.workspaces.section' })
    ).not.toBeInTheDocument();
  });

  it('does not render TeamSection for a member who reaches it via ?section=team', () => {
    tokenState.token = makeToken('member');
    searchState.section = 'team';
    renderSettings();

    expect(
      screen.queryByRole('heading', { name: 'flow.settings.team.section' })
    ).not.toBeInTheDocument();
  });

  // The positive case for ?section=workspaces is deliberately gone. #925 removed
  // WorkspacesSection along with the workspace layer, and SECTION_CONTENT no
  // longer maps the id, so an admin selecting Workspaces now falls through to
  // PlaceholderSection. There is no heading to assert.
  //
  // The three sibling cases above still hold and are still worth keeping: the
  // nav entry is still built for an admin and still withheld from a member, so
  // the #793 gate itself is unchanged.
  //
  // Left alone on purpose: that nav entry is now dead, offering a Workspaces
  // item that renders a placeholder. Removing it means changing buildSections
  // and would break the two sibling assertions, so it belongs in its own change
  // rather than in a test repair.
});

// Flow team membership is read-only here. Dash owns app licenses and role
// assignments, so these tests protect the boundary from accidental mutations.
describe('TeamSection read-only boundary', () => {
  beforeEach(() => {
    flowTeamData.data = [
      { id: '1', name: 'Ada Lovelace', email: 'ada@example.com', apps: ['flow'], isActive: true },
    ];
    grantMutate.mockClear();
    deactivateMutate.mockClear();
    authState.effectiveAccess = undefined;
  });

  it('lists Flow members without mutation controls for a non-admin', () => {
    tokenState.token = makeToken('member');
    renderTeamSection();

    expect(screen.getByText('Ada Lovelace')).toBeInTheDocument();
    expect(
      // These two names no longer resolve to anything: e5d28904f removed the
      // controls and the keys went with them in #1047. That is the point of
      // the assertion, and the string is kept as the record of what must stay
      // absent.
      screen.queryByRole('button', { name: 'flow.settings.team.addMember' })
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.team.removeMemberLabel' })
    ).not.toBeInTheDocument();
  });

  it('keeps mutation controls absent for an admin as well', () => {
    tokenState.token = makeToken('admin');
    authState.effectiveAccess = makeFlowAccess(['flow:team:manage']);
    renderTeamSection();

    expect(screen.getByText('Ada Lovelace')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.team.addMember' })
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.team.removeMemberLabel' })
    ).not.toBeInTheDocument();
  });
});

// The shared EmptyState (title + CTA) now covers both the team-member list
// and the labels list when they have nothing to show.
describe('TeamSection empty state', () => {
  beforeEach(() => {
    flowTeamData.data = [];
    grantMutate.mockClear();
    deactivateMutate.mockClear();
    authState.effectiveAccess = undefined;
  });

  it('shows the empty title with a Refresh CTA for an admin', () => {
    tokenState.token = makeToken('admin');
    authState.effectiveAccess = makeFlowAccess(['flow:team:manage']);
    renderTeamSection();

    expect(screen.getByText('flow.settings.team.noMembersYet')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.team.refreshCta' })).toBeEnabled();
  });

  it('shows the empty title with a Refresh CTA for a non-admin', () => {
    tokenState.token = makeToken('member');
    renderTeamSection();

    expect(screen.getByText('flow.settings.team.noMembersYet')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.team.addMember' })
    ).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.team.refreshCta' })).toBeEnabled();
  });
});

describe('LabelsSection empty state', () => {
  beforeEach(() => {
    labelsData.data = [];
  });

  it('shows the empty title with a CTA that focuses the new-label input', () => {
    renderLabelsSection();

    expect(screen.getByText('flow.settings.labels.emptyList')).toBeInTheDocument();
    const button = screen.getByRole('button', { name: 'flow.settings.labels.newLabel' });
    expect(button).toBeEnabled();
    const input = screen.getByPlaceholderText('flow.settings.labels.namePlaceholder');
    fireEvent.click(button);
    expect(input).toHaveFocus();
  });
});

// #1025: getMe can answer with a role and no effectiveAccess. Reading that as
// "no access" hid the management sections from the people who own the business.
describe('FlowSettings gating on a legacy payload (#1025)', () => {
  beforeEach(() => {
    flowTeamData.data = [];
    searchState.section = undefined;
    authState.effectiveAccess = undefined;
    authState.user = undefined;
  });

  it.each(['owner', 'admin'])(
    'shows Team and Workspaces to %s when the payload carries no effective access',
    (role) => {
      authState.user = { role };
      renderSettings();

      expect(screen.getByRole('button', { name: 'common.settings.team' })).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: 'flow.settings.workspaces.section' })
      ).toBeInTheDocument();
    }
  );

  it('still hides them from a member on the same legacy payload', () => {
    authState.user = { role: 'member' };
    renderSettings();

    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.workspaces.section' })
    ).not.toBeInTheDocument();
  });

  it('hides them when there is neither effective access nor a role', () => {
    renderSettings();

    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
  });

  it('lets effective access override a privileged legacy role rather than widening it', () => {
    authState.user = { role: 'owner' };
    authState.effectiveAccess = makeFlowAccess(['flow:projects:view']);
    renderSettings();

    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
  });
});
