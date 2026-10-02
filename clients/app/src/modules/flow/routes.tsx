import React, { Suspense } from 'react';
import { createRoute, Outlet, Navigate, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
const FlowLayout = React.lazy(() => import('./layouts/FlowLayout'));

import PreviewRouteUnavailable from './shared/components/PreviewRouteUnavailable';
import RouteNotFound from './shared/components/RouteNotFound';
import { settingsSearchSchema } from './shared/schemas/settings';
import { myDaySearchSchema } from './shared/schemas/time';
import { taskDrawerSearchSchema } from './shared/schemas/taskDrawer';
import { useTranslation } from 'react-i18next';
import { useUserStore } from '../../shared/stores/userStore';
import { getFlowHomePath } from './navigation';
// The preview-route fallback is eager so a disabled deep link can explain its state
// without waiting for a lazy route component to load.

const ProjectBoard = React.lazy(() => import('./features/projects/ProjectBoard'));
const ProjectList = React.lazy(() => import('./features/projects/ProjectList'));
const ProjectOverview = React.lazy(() => import('./features/projects/ProjectOverview'));
const ProjectSettings = React.lazy(() => import('./features/projects/ProjectSettings'));
const ProjectTimeline = React.lazy(() => import('./features/projects/ProjectTimeline'));
const ProjectCalendar = React.lazy(() => import('./features/projects/ProjectCalendar'));
const ProjectDashboard = React.lazy(() => import('./features/projects/ProjectDashboard'));
const Inbox = React.lazy(() => import('./features/inbox/Inbox'));
const MyDay = React.lazy(() => import('./features/time/MyDay'));
const LeaveScreen = React.lazy(() => import('./features/leave/LeaveScreen'));
const DirectoryScreen = React.lazy(() => import('./features/directory/DirectoryScreen'));
const PresenceScreen = React.lazy(() => import('./features/presence/PresenceScreen'));
const TeamView = React.lazy(() => import('./features/time/TeamView'));
const TimeReports = React.lazy(() => import('./features/time/TimeReports'));
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
const Start = React.lazy(() => import('./features/Start'));
const MyIssues = React.lazy(() => import('./features/tasks/MyIssues'));
const Archive = React.lazy(() => import('./features/archive/Archive'));

const RouteLoading = (): React.JSX.Element => {
  const { t } = useTranslation();
  return (
    <div
      style={{
        padding: '32px',
        color: 'var(--flow-text-muted)',
        fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
        fontSize: 'var(--text-flow-body)',
      }}
    >
      {t('flow.shell.loading')}
    </div>
  );
};

function withLayout(element: React.ReactNode): React.ReactNode {
  return (
    <ErrorBoundary>
      <Suspense fallback={<RouteLoading />}>{element}</Suspense>
    </ErrorBoundary>
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

export function createFlowRoutes<TParent extends AnyRoute>(parent: TParent) {
  const rootRoute = createRoute({
    getParentRoute: () => parent,
    path: 'flow',
    validateSearch: taskDrawerSearchSchema,
    component: () => (
      <Suspense fallback={<RouteLoading />}>
        <FlowLayout>
          <Outlet />
        </FlowLayout>
      </Suspense>
    ),
  });

  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => {
      const user = useUserStore((state) => state.user);
      return <Navigate to={getFlowHomePath(user)} replace search={rootRoute.useSearch()} />;
    },
  });

  const startRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/start',
    component: () => withLayout(<Start />),
  });

  const projectBoardRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId',
    component: () => withLayout(<ProjectBoard />),
  });

  const projectBoardExplicitRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/board',
    component: () => withLayout(<ProjectBoard />),
  });

  const projectListRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/list',
    component: () => withLayout(<ProjectList />),
  });

  const projectOverviewRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/overview',
    component: () => previewRoute(<ProjectOverview />),
  });

  const projectTimelineRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/timeline',
    component: () => previewRoute(<ProjectTimeline />),
  });

  const projectCalendarRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/calendar',
    component: () => withLayout(<ProjectCalendar />),
  });

  const projectDashboardRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/dashboard',
    component: () => previewRoute(<ProjectDashboard />),
  });

  const projectSettingsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/p/$projectId/settings',
    component: () => withLayout(<ProjectSettings />),
  });

  const taskDetailRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/task/$id',
    component: () => withLayout(<TaskDetail />),
  });

  const AllProjects = React.lazy(() => import('./features/projects-v2/AllProjects'));
  const Profile = React.lazy(() => import('./features/profile-v2/Profile'));

  // "My tasks" shows every task assigned to the user across all workspaces and
  // projects (it is not workspace-scoped): a single cross-project board.
  const myIssuesRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/my-issues',
    component: () => withLayout(<MyIssues />),
  });

  const allProjectsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/projects',
    component: () => withLayout(<AllProjects />),
  });

  const profileRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/u/$userId',
    component: () => {
      const { userId } = profileRoute.useParams();
      return withLayout(<Profile userId={userId} />);
    },
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
    validateSearch: myDaySearchSchema,
    component: () => withLayout(<MyDay />),
  });

  const leaveRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/leave',
    component: () => withLayout(<LeaveScreen />),
  });

  const teamViewRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/time/team',
    component: () => withLayout(<TeamView />),
  });

  const timeReportsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/time/reports',
    component: () => withLayout(<TimeReports />),
  });

  const directoryRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/directory',
    component: () => withLayout(<DirectoryScreen />),
  });

  const presenceRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/presence',
    component: () => withLayout(<PresenceScreen />),
  });

  const settingsRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/settings',
    validateSearch: settingsSearchSchema,
    component: () => <Navigate to="/settings/flow" replace search={settingsRoute.useSearch()} />,
  });

  const catchAllRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '$',
    component: RouteNotFound,
  });

  return rootRoute.addChildren([
    indexRoute,
    startRoute,

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
    allProjectsRoute,
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
    profileRoute,
    leaveRoute,
    teamViewRoute,
    timeReportsRoute,
    directoryRoute,
    presenceRoute,
    settingsRoute,
    catchAllRoute,
  ]);
}
