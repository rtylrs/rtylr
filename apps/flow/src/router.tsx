import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  Navigate,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
import { unwrap } from '@voxire/api';
import { AppGuard } from '@voxire/auth';
import FlowLayout from './layouts/FlowLayout';
import { PERSONAL_WORKSPACE_ID } from './shared/lib/personalWorkspace';
import { safe } from './shared/lib/api';
import PreviewRouteUnavailable from './shared/components/PreviewRouteUnavailable';
import { settingsSearchSchema } from './shared/schemas/settings';
// The preview-route fallback is eager so a disabled deep link can explain its state
// without waiting for a lazy route component to load.

const WorkspaceDashboard = React.lazy(
  () => import('./features/workspace-dashboard/WorkspaceDashboard')
);
const ProjectBoard = React.lazy(() => import('./features/projects/ProjectBoard'));
const ProjectList = React.lazy(() => import('./features/projects/ProjectList'));
const ProjectOverview = React.lazy(() => import('./features/projects/ProjectOverview'));
const ProjectSettings = React.lazy(() => import('./features/projects/ProjectSettings'));
const ProjectTimeline = React.lazy(() => import('./features/projects/ProjectTimeline'));
const ProjectCalendar = React.lazy(() => import('./features/projects/ProjectCalendar'));
const ProjectDashboard = React.lazy(() => import('./features/projects/ProjectDashboard'));
const Inbox = React.lazy(() => import('./features/inbox/Inbox'));
const FlowSettings = React.lazy(() => import('./features/settings/FlowSettings'));
const MyDay = React.lazy(() => import('./features/time/MyDay'));
const TeamView = React.lazy(() => import('./features/time/TeamView'));
const TaskDetail = React.lazy(() => import('./features/tasks/TaskDetail'));
const Timeline = React.lazy(() => import('./features/timeline/Timeline'));
const CalendarView = React.lazy(() => import('./features/calendar/Calendar'));
const Milestones = React.lazy(() => import('./features/milestones/Milestones'));
const Cycles = React.lazy(() => import('./features/cycles/Cycles'));
const Views = React.lazy(() => import('./features/views/Views'));
const Roadmap = React.lazy(() => import('./features/roadmap/Roadmap'));
const Insights = React.lazy(() => import('./features/insights/Insights'));
const Team = React.lazy(() => import('./features/team/Team'));
const Triage = React.lazy(() => import('./features/triage/Triage'));
const FocusToday = React.lazy(() => import('./features/focus-today/FocusToday'));
const MyIssues = React.lazy(() => import('./features/tasks/MyIssues'));
const Archive = React.lazy(() => import('./features/archive/Archive'));
const WorkspaceManager = React.lazy(() => import('./features/workspace-manager/WorkspaceManager'));

function withLayout(element: React.ReactNode): React.ReactNode {
  return (
    <FlowLayout>
      <ErrorBoundary>
        <Suspense
          fallback={
            <div
              style={{
                padding: '32px',
                color: 'var(--flow-text-muted)',
                fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
                fontSize: '13px',
              }}
            >
              Loading...
            </div>
          }
        >
          {element}
        </Suspense>
      </ErrorBoundary>
    </FlowLayout>
  );
}

function previewRoute(element: React.ReactNode): React.ReactNode {
  return (
    <AppFlagGuard
      flag={FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES}
      fallback={withLayout(<PreviewRouteUnavailable />)}
    >
      {withLayout(element)}
    </AppFlagGuard>
  );
}

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="flow"
        fetchUser={(): Promise<Record<string, unknown>> =>
          unwrap(safe.get<Record<string, unknown>>('/hub/user/v1/users/me'))
        }
      >
        <Outlet />
      </AppGuard>
    </ErrorBoundary>
  ),
});

// Landing on Flow drops the user on their global "My tasks" view (every task
// assigned to them across all workspaces and projects), not a workspace board.
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => <Navigate to="/my-issues" replace />,
});

const workspaceDashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId',
  component: () => withLayout(<WorkspaceDashboard />),
});

const workspaceMembersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/members',
  component: () => withLayout(<WorkspaceManager />),
});

const projectBoardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId',
  component: () => withLayout(<ProjectBoard />),
});

const projectBoardExplicitRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/board',
  component: () => withLayout(<ProjectBoard />),
});

const projectListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/list',
  component: () => withLayout(<ProjectList />),
});

const projectOverviewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/overview',
  component: () => previewRoute(<ProjectOverview />),
});

const projectTimelineRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/timeline',
  component: () => previewRoute(<ProjectTimeline />),
});

const projectCalendarRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/calendar',
  component: () => withLayout(<ProjectCalendar />),
});

const projectDashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/dashboard',
  component: () => previewRoute(<ProjectDashboard />),
});

const projectSettingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/w/$workspaceId/p/$projectId/settings',
  component: () => withLayout(<ProjectSettings />),
});

const taskDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/task/$id',
  component: () => withLayout(<TaskDetail />),
});

// "My tasks" shows every task assigned to the user across all workspaces and
// projects (it is not workspace-scoped): a single cross-project board.
const myIssuesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/my-issues',
  component: () => withLayout(<MyIssues />),
});

const todayRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/today',
  component: () => withLayout(<FocusToday />),
});

const timelineRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/timeline',
  component: () => previewRoute(<Timeline />),
});

const calendarRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/calendar',
  component: () => withLayout(<CalendarView />),
});

const milestonesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/milestones',
  component: () => withLayout(<Milestones />),
});

const cyclesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/cycles',
  component: () => withLayout(<Cycles />),
});

const viewsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/views',
  component: () => withLayout(<Views />),
});

const roadmapRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/roadmap',
  component: () => previewRoute(<Roadmap />),
});

const insightsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/insights',
  component: () => previewRoute(<Insights />),
});

const teamRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/team',
  component: () => previewRoute(<Team />),
});

const triageRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/triage',
  component: () => withLayout(<Triage />),
});

const archiveRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/archive',
  component: () => withLayout(<Archive />),
});

const inboxRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/inbox',
  component: () => withLayout(<Inbox />),
});

const myDayRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/time',
  component: () => withLayout(<MyDay />),
});

const teamViewRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/time/team',
  component: () => withLayout(<TeamView />),
});

const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  validateSearch: settingsSearchSchema,
  component: () => withLayout(<FlowSettings />),
});

const catchAllRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '*',
  component: () => (
    <Navigate to="/w/$workspaceId" params={{ workspaceId: PERSONAL_WORKSPACE_ID }} replace />
  ),
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  workspaceDashboardRoute,
  workspaceMembersRoute,
  projectBoardRoute,
  projectBoardExplicitRoute,
  projectListRoute,
  projectOverviewRoute,
  projectTimelineRoute,
  projectCalendarRoute,
  projectDashboardRoute,
  projectSettingsRoute,
  taskDetailRoute,
  myIssuesRoute,
  todayRoute,
  timelineRoute,
  calendarRoute,
  milestonesRoute,
  cyclesRoute,
  viewsRoute,
  roadmapRoute,
  insightsRoute,
  teamRoute,
  triageRoute,
  archiveRoute,
  inboxRoute,
  myDayRoute,
  teamViewRoute,
  settingsRoute,
  catchAllRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
