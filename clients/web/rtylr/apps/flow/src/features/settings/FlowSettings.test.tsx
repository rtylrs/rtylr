import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return { ...actual, useTranslation: () => ({ t: (key: string) => key }) };
});
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('@voxire/config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/config')>();
  return {
    ...actual,
    useFeatureFlag: () => false,
    FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
  };
});
vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return { ...actual, unwrap: () => Promise.resolve({}) };
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
    memberType: string;
    isOwner: boolean;
    hasFlowAccess: boolean;
    isActive: boolean;
  }>,
}));
vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: flowTeamData.data, isLoading: false }),
  useFlowDeactivateMember: () => ({ mutate: vi.fn(), isPending: false }),
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

const authState = vi.hoisted(() => ({ effectiveAccess: undefined as EffectiveAccess | undefined }));
vi.mock('@voxire/auth', () => ({
  getCookie: () => undefined,
  setCookie: vi.fn(),
  getDefaultCookieDomain: () => undefined,
  useUserStore: (selector: (state: typeof authState) => unknown): unknown => selector(authState),
}));

const searchState = vi.hoisted(() => ({ section: undefined as string | undefined }));
vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => searchState }),
}));

import FlowSettings, { TeamSection, LabelsSection } from './FlowSettings';

function makeFlowAccess(
  permissions: string[] = [],
  flags: Partial<Pick<EffectiveAccess, 'isOwner' | 'isAdmin'>> = {}
): EffectiveAccess {
  return {
    businessId: '10',
    dashPermissions: [],
    subscribedApps: ['flow'],
    licenses: [],
    apps: [
      {
        appId: 'flow',
        licensed: true,
        hasAccess: true,
        permissions,
        permissionScopes: Object.fromEntries(permissions.map((permission) => [permission, []])),
        sources: ['explicit'],
      },
    ],
    isOwner: flags.isOwner ?? false,
    isAdmin: flags.isAdmin ?? false,
    authorizationVersion: 1,
  };
}

function renderWith(component: React.ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={queryClient}>{component}</QueryClientProvider>);
}

beforeEach(() => {
  flowTeamData.data = [];
  labelsData.data = [];
  createLabelMutate.mockClear();
  searchState.section = undefined;
  authState.effectiveAccess = undefined;
});

describe('Flow settings permissions', () => {
  it('gives an administrator full access to management sections', () => {
    authState.effectiveAccess = makeFlowAccess([], { isAdmin: true });
    renderWith(<FlowSettings />);

    expect(screen.getByRole('button', { name: 'common.settings.team' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.time.section' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.bin.section' })).toBeInTheDocument();
  });

  it('shows personal settings to members and gates business actions by permission', () => {
    authState.effectiveAccess = makeFlowAccess(['flow:team:manage']);
    renderWith(<FlowSettings />);

    expect(screen.getByRole('button', { name: 'common.settings.team' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.time.section' })).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.bin.section' })
    ).not.toBeInTheDocument();
  });

  it('keeps personal settings available when effective access is absent', () => {
    renderWith(<FlowSettings />);
    expect(screen.queryByRole('button', { name: 'common.settings.team' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.settings.time.section' })).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.bin.section' })
    ).not.toBeInTheDocument();
  });

  it('validates a requested section against the same effective permission gate', () => {
    authState.effectiveAccess = makeFlowAccess([]);
    searchState.section = 'team';
    renderWith(<FlowSettings />);
    expect(
      screen.queryByRole('heading', { name: 'flow.settings.team.section' })
    ).not.toBeInTheDocument();
  });

  it('uses Notifications as the default section for both members and admins', () => {
    authState.effectiveAccess = makeFlowAccess([], { isAdmin: true });
    renderWith(<FlowSettings />);
    expect(
      screen.getByText('flow.settings.notifications.items.taskAssigned.label')
    ).toBeInTheDocument();
  });
});

describe('Flow team section', () => {
  it('shows only members with resolved Flow access and has no access mutation controls', () => {
    flowTeamData.data = [
      {
        id: '1',
        name: 'Ada Lovelace',
        email: 'ada@example.com',
        memberType: 'member',
        isOwner: false,
        hasFlowAccess: true,
        isActive: true,
      },
      {
        id: '2',
        name: 'Grace Hopper',
        email: 'grace@example.com',
        memberType: 'member',
        isOwner: false,
        hasFlowAccess: false,
        isActive: true,
      },
    ];
    renderWith(<TeamSection />);

    expect(screen.getByText('Ada Lovelace')).toBeInTheDocument();
    expect(screen.queryByText('Grace Hopper')).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.settings.team.addMember' })
    ).not.toBeInTheDocument();
  });

  it('shows an empty state when no member has Flow access', () => {
    flowTeamData.data = [
      {
        id: '2',
        name: 'Grace Hopper',
        email: 'grace@example.com',
        memberType: 'member',
        isOwner: false,
        hasFlowAccess: false,
        isActive: true,
      },
    ];
    renderWith(<TeamSection />);
    expect(screen.getByText('flow.settings.team.noMembersYet')).toBeInTheDocument();
  });
});

describe('Labels section', () => {
  it('focuses the new-label input from the empty-state action', () => {
    renderWith(<LabelsSection />);
    const input = screen.getByPlaceholderText('flow.settings.labels.namePlaceholder');
    fireEvent.click(screen.getByRole('button', { name: 'flow.settings.labels.newLabel' }));
    expect(input).toHaveFocus();
  });
});
