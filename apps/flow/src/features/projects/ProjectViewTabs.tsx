import React from 'react';
import { Link, useLocation } from '@tanstack/react-router';
import {
  Activity,
  LayoutGrid,
  List,
  GanttChart,
  Calendar,
  BarChart3,
  Settings as SettingsIcon,
  Plus,
} from 'lucide-react';
import { useFeatureFlag, FLAG_KEYS } from '@voxire/config';

interface ProjectViewTabsProps {
  project: { name: string; color: string };
  workspaceId: string;
  projectId: string;
  taskCount: number;
  onNewTask: () => void;
  rightSlot?: React.ReactNode;
}

interface ViewDef {
  key: string;
  label: string;
  icon: React.ComponentType<{ size?: number }>;
  path: (wid: string, pid: string) => string;
  match: (path: string) => boolean;
  // Route is wrapped in AppFlagGuard(APP_FLOW_PREVIEW_ROUTES) in router.tsx -
  // keep this in sync with which routes are gated there.
  previewOnly?: boolean;
}

const VIEWS: ViewDef[] = [
  {
    key: 'overview',
    label: 'Overview',
    icon: Activity,
    path: (w, p) => `/w/${w}/p/${p}/overview`,
    match: (p) => p.endsWith('/overview'),
    previewOnly: true,
  },
  {
    key: 'board',
    label: 'Board',
    icon: LayoutGrid,
    path: (w, p) => `/w/${w}/p/${p}`,
    // The default project route is board - matches anything that isn't
    // explicitly another view.
    match: (p) => /\/w\/[^/]+\/p\/[^/]+$/.test(p) || p.endsWith('/board'),
  },
  {
    key: 'list',
    label: 'List',
    icon: List,
    path: (w, p) => `/w/${w}/p/${p}/list`,
    match: (p) => p.endsWith('/list'),
  },
  {
    key: 'timeline',
    label: 'Timeline',
    icon: GanttChart,
    path: (w, p) => `/w/${w}/p/${p}/timeline`,
    match: (p) => p.endsWith('/timeline'),
    previewOnly: true,
  },
  {
    key: 'calendar',
    label: 'Calendar',
    icon: Calendar,
    path: (w, p) => `/w/${w}/p/${p}/calendar`,
    match: (p) => p.endsWith('/calendar'),
  },
  {
    key: 'dashboard',
    label: 'Dashboard',
    icon: BarChart3,
    path: (w, p) => `/w/${w}/p/${p}/dashboard`,
    match: (p) => p.endsWith('/dashboard'),
    previewOnly: true,
  },
];

/**
 * Shared project header. Renders the project name + task count on the left
 * and the full view-tab switcher + primary "New task" CTA on the right.
 * Every project view (Overview, Board, List, Timeline, Calendar, Dashboard)
 * renders this so the navigation is stable across views.
 */
const ProjectViewTabs: React.FC<ProjectViewTabsProps> = ({
  project,
  workspaceId,
  projectId,
  taskCount,
  onNewTask,
  rightSlot,
}) => {
  const { pathname } = useLocation();
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const visibleViews = VIEWS.filter((v) => !v.previewOnly || showPreviewRoutes);

  return (
    <div
      className="flex flex-col gap-2 px-4 py-2.5 shrink-0 sm:flex-row sm:items-center sm:justify-between sm:px-5 sm:py-3"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <div className="flex min-w-0 items-center gap-2.5">
        <span className="h-2.5 w-2.5 rounded-full shrink-0" style={{ background: project.color }} />
        <h1 className="text-[14px] font-semibold truncate" style={{ color: 'var(--flow-text)' }}>
          {project.name}
        </h1>
        <span
          className="text-[10px] px-1.5 py-0.5 rounded-full font-semibold shrink-0"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
        >
          {taskCount}
        </span>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        {rightSlot}

        {/* View switcher */}
        <div
          className="flex items-center rounded-md p-0.5"
          style={{
            background: 'var(--flow-surface)',
            border: '1px solid var(--flow-border)',
          }}
        >
          {visibleViews.map((v) => {
            const active = v.match(pathname);
            const Icon = v.icon;
            return (
              <Link
                key={v.key}
                to={v.path(workspaceId, projectId) as string}
                className={[
                  'flex items-center gap-1.5 px-2 py-1 rounded text-[11px] font-medium transition-colors',
                  active ? '' : 'hover:bg-[var(--flow-hover)]',
                ].join(' ')}
                style={
                  active
                    ? { background: 'var(--accent-muted)', color: 'var(--accent)' }
                    : { color: 'var(--flow-text-muted)' }
                }
              >
                <Icon size={11} />
                <span className="hidden md:inline">{v.label}</span>
              </Link>
            );
          })}
        </div>

        <Link
          to={`/w/${workspaceId}/p/${projectId}/settings` as string}
          className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)] transition-colors"
          style={{ color: 'var(--flow-text-dim)' }}
          aria-label="Project settings"
        >
          <SettingsIcon size={13} />
        </Link>

        {/* Primary CTA. White text on teal - the old near-black-on-teal
            combo was failing WCAG on anything but the darkest teal. */}
        <button
          onClick={onNewTask}
          className="flex items-center gap-1.5 h-7 px-2.5 rounded text-[11px] font-semibold transition-opacity hover:opacity-90"
          style={{ background: 'var(--accent)', color: '#ffffff' }}
        >
          <Plus size={12} strokeWidth={2.5} /> New task
        </button>
      </div>
    </div>
  );
};

export default ProjectViewTabs;
