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
import { useTranslation } from 'react-i18next';
import { useFeatureFlag, FLAG_KEYS } from '@voxire/config';
import ProjectIcon from '../../shared/components/ProjectIcon';

interface ProjectViewTabsProps {
  project: { name: string; color: string; icon?: string };
  projectId: string;
  taskCount: number;
  onNewTask: () => void;
  rightSlot?: React.ReactNode;
}

type ProjectViewPath =
  | '/flow/p/$projectId'
  | '/flow/p/$projectId/overview'
  | '/flow/p/$projectId/list'
  | '/flow/p/$projectId/timeline'
  | '/flow/p/$projectId/calendar'
  | '/flow/p/$projectId/dashboard';

interface ViewDef {
  key: string;
  /**
   * A translation key, not a label. This array is module scope, so it is built
   * before any hook can run and cannot call `t` itself; the tab resolves the
   * key at render.
   */
  labelKey: string;
  icon: React.ComponentType<{ size?: number }>;
  // A literal from the route tree, not a builder: `<Link to>` is typed against
  // the tree, so an interpolated path needs a cast, and the cast is what let
  // every one of these links keep pointing at /w/... after the routes moved.
  to: ProjectViewPath;
  match: (path: string) => boolean;
  // Route is wrapped in AppFlagGuard(APP_FLOW_PREVIEW_ROUTES) in router.tsx -
  // keep this in sync with which routes are gated there.
  previewOnly?: boolean;
}

const VIEWS: ViewDef[] = [
  {
    key: 'overview',
    labelKey: 'flow.projectViewTabs.overview',
    icon: Activity,
    to: '/flow/p/$projectId/overview',
    match: (p) => p.endsWith('/overview'),
    previewOnly: true,
  },
  {
    key: 'board',
    labelKey: 'flow.projectViewTabs.board',
    icon: LayoutGrid,
    to: '/flow/p/$projectId',
    // The default project route is board - matches anything that isn't
    // explicitly another view.
    match: (p) => /^\/p\/[^/]+$/.test(p) || p.endsWith('/board'),
  },
  {
    key: 'list',
    labelKey: 'flow.projectViewTabs.list',
    icon: List,
    to: '/flow/p/$projectId/list',
    match: (p) => p.endsWith('/list'),
  },
  {
    key: 'timeline',
    labelKey: 'flow.shell.timeline',
    icon: GanttChart,
    to: '/flow/p/$projectId/timeline',
    match: (p) => p.endsWith('/flow/timeline'),
    previewOnly: true,
  },
  {
    key: 'calendar',
    labelKey: 'flow.shell.calendar',
    icon: Calendar,
    to: '/flow/p/$projectId/calendar',
    match: (p) => p.endsWith('/flow/calendar'),
  },
  {
    key: 'dashboard',
    labelKey: 'flowNav.dashboard',
    icon: BarChart3,
    to: '/flow/p/$projectId/dashboard',
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
  projectId,
  taskCount,
  onNewTask,
  rightSlot,
}) => {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const showPreviewRoutes = useFeatureFlag(FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES);
  const visibleViews = VIEWS.filter((v) => !v.previewOnly || showPreviewRoutes);

  return (
    <div
      className="flex flex-col gap-2 px-4 py-2.5 shrink-0 sm:flex-row sm:items-center sm:justify-between sm:px-5 sm:py-3"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      <div className="flex min-w-0 items-center gap-2.5">
        {/* The project's own glyph, tinted with its colour, rather than a
            plain dot. G-C5 puts the colour in three places: the sidebar dot,
            this header icon and the chip. The dot said nothing the sidebar was
            not already saying (#1085). 20px, per G-I1's header size. */}
        <ProjectIcon
          icon={project.icon}
          size={20}
          className="shrink-0"
          style={{ color: project.color }}
        />
        <h1
          className="text-flow-section font-semibold truncate"
          style={{ color: 'var(--flow-text)' }}
        >
          {project.name}
        </h1>
        <span
          className="text-flow-meta px-1.5 py-0.5 rounded-full font-semibold shrink-0"
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
                to={v.to}
                params={{ projectId }}
                className={[
                  'flex items-center gap-1.5 px-2 py-1 rounded-sm text-flow-meta font-medium transition-colors',
                  active ? '' : 'hover:bg-[var(--flow-hover)]',
                ].join(' ')}
                style={
                  active
                    ? { background: 'var(--accent-muted)', color: 'var(--accent-on-muted)' }
                    : { color: 'var(--flow-text-muted)' }
                }
              >
                <Icon size={11} />
                <span className="hidden md:inline">{t(v.labelKey)}</span>
              </Link>
            );
          })}
        </div>

        <Link
          to="/flow/p/$projectId/settings"
          params={{ projectId }}
          className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] transition-colors"
          style={{ color: 'var(--flow-text-dim)' }}
          aria-label={t('flowNav.settings')}
        >
          <SettingsIcon size={13} />
        </Link>

        {/* Primary CTA. White text on teal - the old near-black-on-teal
            combo was failing WCAG on anything but the darkest teal. */}
        <button
          type="button"
          onClick={onNewTask}
          className="flex items-center gap-1.5 h-7 px-2.5 rounded-sm text-flow-meta font-semibold transition-opacity hover:opacity-90"
          style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
        >
          <Plus size={12} strokeWidth={2.5} /> New task
        </button>
      </div>
    </div>
  );
};

export default ProjectViewTabs;
