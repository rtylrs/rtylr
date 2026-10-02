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
vi.mock('../../../../shared/stores/userStore', () => ({
  getCookie: () => undefined,
  setCookie: vi.fn(),
  getDefaultCookieDomain: () => undefined,
  useUserStore: (selector: (state: { user: typeof authState }) => unknown): unknown =>
    selector({ user: authState }),
}));

const searchState = vi.hoisted(() => ({ section: undefined as string | undefined }));
vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => searchState }),
}));

import FlowSettings, { LabelsSection } from './FlowSettings';

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
  labelsData.data = [];
  createLabelMutate.mockClear();
  searchState.section = undefined;
  authState.effectiveAccess = undefined;
});

describe('Flow settings permissions', () => {
  it('renders the panel chosen in shared settings navigation', () => {
    searchState.section = 'labels';
    authState.effectiveAccess = makeFlowAccess([], { isAdmin: true });
    renderWith(<FlowSettings />);
    expect(screen.getByPlaceholderText('flow.settings.labels.namePlaceholder')).toBeVisible();
    expect(screen.queryByText('flow.settings.notifications.items.taskAssigned.label')).toBeNull();
  });
  it('ignores a management panel without its permission', () => {
    searchState.section = 'bin';
    authState.effectiveAccess = makeFlowAccess([]);
    renderWith(<FlowSettings />);
    expect(screen.getByText('flow.settings.notifications.items.taskAssigned.label')).toBeVisible();
  });
  it('validates a requested section against the same effective permission gate', () => {
    authState.effectiveAccess = makeFlowAccess([]);
    searchState.section = 'unknown';
    renderWith(<FlowSettings />);
    expect(screen.queryByRole('heading', { name: 'unknown' })).not.toBeInTheDocument();
  });

  it('uses Notifications as the default section for both members and admins', () => {
    authState.effectiveAccess = makeFlowAccess([], { isAdmin: true });
    renderWith(<FlowSettings />);
    expect(
      screen.getByText('flow.settings.notifications.items.taskAssigned.label')
    ).toBeInTheDocument();
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
