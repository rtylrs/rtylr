// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';

vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (k: string) => k }) }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children: React.ReactNode }) => <a href="/">{children}</a>,
  useNavigate: () => vi.fn(),
  useLocation: () => ({ pathname: '/' }),
}));

import { SidebarNavigation } from './SidebarNavigation';
import { createDefaultSidebarPreferences } from '../lib/sidebarPreferences';
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

function renderNav(projects: ApiProject[]) {
  return render(
    <SidebarNavigation
      preferences={createDefaultSidebarPreferences()}
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
