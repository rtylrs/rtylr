// @vitest-environment happy-dom
import React from 'react';
import { beforeAll, describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import path from 'node:path';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children: React.ReactNode }) => <a href="/">{children}</a>,
  useNavigate: () => vi.fn(),
  useLocation: () => ({ pathname: '/' }),
}));

import { SidebarNavigation } from './SidebarNavigation';
import {
  createDefaultSidebarPreferences,
  recordRecentSidebarPage,
} from '../lib/sidebarPreferences';
import type { ApiProject } from '../services/flowApi';

function project(over: Partial<ApiProject> = {}): ApiProject {
  return {
    id: '1',
    name: 'Web rebuild',
    color: '#64C5B9',
    icon: 'rocket',
    status: 'active',
    taskCount: 0,
    completedCount: 0,
    ...over,
  } as unknown as ApiProject;
}

function renderNav(projects: ApiProject[], preferences = createDefaultSidebarPreferences()) {
  return render(
    <SidebarNavigation
      preferences={preferences}
      update={vi.fn()}
      showPreviewRoutes={false}
      unreadNotifications={0}
      myIssuesCount={0}
      todayFocusCount={0}
      projects={projects}
      onNewProject={vi.fn()}
      sidebarRef={{ current: null }}
    />
  );
}

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

it('translates recent builtin routes from current labels and preserves project names', () => {
  let preferences = createDefaultSidebarPreferences();
  preferences = recordRecentSidebarPage(preferences, '/today', 'nav.focusToday');
  preferences = recordRecentSidebarPage(preferences, '/inbox', 'nav.inbox');
  preferences = recordRecentSidebarPage(preferences, '/projects/41', 'Launch project');
  renderNav([], preferences);

  expect(screen.queryByText('nav.focusToday')).not.toBeInTheDocument();
  expect(screen.queryByText('nav.inbox')).not.toBeInTheDocument();
  expect(screen.getAllByText('Today')).toHaveLength(2);
  expect(screen.getAllByText('Inbox')).toHaveLength(2);
  expect(screen.getByText('Launch project')).toBeInTheDocument();
});

/**
 * A project row showed its open-task count unconditionally, so a project with
 * everything finished rendered a bare "0" (#1097, D-058). A zero there reads
 * as a number worth noting rather than as nothing left to do.
 */
describe('sidebar project counts', () => {
  it('hides the count when nothing is open', () => {
    renderNav([project({ taskCount: 6, completedCount: 6 })]);

    expect(screen.getByText('Web rebuild')).toBeInTheDocument();
    expect(screen.queryByText('0')).not.toBeInTheDocument();
  });

  it('shows the count when work remains', () => {
    renderNav([project({ taskCount: 6, completedCount: 2 })]);
    expect(screen.getByText('4')).toBeInTheDocument();
  });

  it('renders the count with tabular figures, so the column does not shift', () => {
    renderNav([project({ taskCount: 6, completedCount: 2 })]);
    expect(screen.getByText('4').className).toContain('tabular-nums');
  });

  it('hides it for a project with no tasks at all, not just a finished one', () => {
    renderNav([project({ taskCount: 0, completedCount: 0 })]);
    expect(screen.queryByText('0')).not.toBeInTheDocument();
  });
});
