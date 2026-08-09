import React, { Suspense } from 'react';
import { Navigate } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import FlowLayout from '../layouts/FlowLayout';
import { PERSONAL_WORKSPACE_ID } from '../shared/lib/personalWorkspace';
import { ErrorBoundary } from '@voxire/ui';

const WorkspaceDashboard = React.lazy(
  () => import('../features/workspace-dashboard/WorkspaceDashboard')
);
const ProjectBoard = React.lazy(() => import('../features/projects/ProjectBoard'));
const ProjectList = React.lazy(() => import('../features/projects/ProjectList'));
const ProjectOverview = React.lazy(() => import('../features/projects/ProjectOverview'));
const ProjectSettings = React.lazy(() => import('../features/projects/ProjectSettings'));
const ProjectTimeline = React.lazy(() => import('../features/projects/ProjectTimeline'));
const ProjectCalendar = React.lazy(() => import('../features/projects/ProjectCalendar'));
const ProjectDashboard = React.lazy(() => import('../features/projects/ProjectDashboard'));
const Inbox = React.lazy(() => import('../features/inbox/Inbox'));
const FlowSettings = React.lazy(() => import('../features/settings/FlowSettings'));
const TaskDetail = React.lazy(() => import('../features/tasks/TaskDetail'));
const MyIssues = React.lazy(() => import('../features/tasks/MyIssues'));
const Timeline = React.lazy(() => import('../features/timeline/Timeline'));
const CalendarView = React.lazy(() => import('../features/calendar/Calendar'));
const Milestones = React.lazy(() => import('../features/milestones/Milestones'));
const Cycles = React.lazy(() => import('../features/cycles/Cycles'));
const Views = React.lazy(() => import('../features/views/Views'));
const Roadmap = React.lazy(() => import('../features/roadmap/Roadmap'));
const Insights = React.lazy(() => import('../features/insights/Insights'));
const Team = React.lazy(() => import('../features/team/Team'));
const Triage = React.lazy(() => import('../features/triage/Triage'));
const FocusToday = React.lazy(() => import('../features/focus-today/FocusToday'));
const Archive = React.lazy(() => import('../features/archive/Archive'));
const WorkspaceManager = React.lazy(() => import('../features/workspace-manager/WorkspaceManager'));

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
              Loading…
            </div>
          }
        >
          {element}
        </Suspense>
      </ErrorBoundary>
    </FlowLayout>
  );
}

/**
 * Stub-route gate. The wrapped route renders only when the
 * `app.flow.previewRoutes` flag is on (off in prod). Otherwise the
 * user is bounced back to their workspace dashboard so they don't see
 * UI-only screens that have no data behind them yet.
 */
function previewRoute(element: React.ReactNode): React.ReactNode {
  return (
    <AppFlagGuard
      flag={FLAG_KEYS.APP_FLOW_PREVIEW_ROUTES}
      fallback={
        <Navigate to="/w/$workspaceId" params={{ workspaceId: PERSONAL_WORKSPACE_ID }} replace />
      }
    >
      {withLayout(element)}
    </AppFlagGuard>
  );
}

const appRoutes = [
  {
    path: '/',
    element: (
      <Navigate to="/w/$workspaceId" params={{ workspaceId: PERSONAL_WORKSPACE_ID }} replace />
    ),
  },
  { path: '/w/:workspaceId', element: withLayout(<WorkspaceDashboard />) },
  { path: '/w/:workspaceId/members', element: withLayout(<WorkspaceManager />) },
  { path: '/w/:workspaceId/p/:projectId', element: withLayout(<ProjectBoard />) },
  { path: '/w/:workspaceId/p/:projectId/board', element: withLayout(<ProjectBoard />) },
  { path: '/w/:workspaceId/p/:projectId/list', element: withLayout(<ProjectList />) },
  { path: '/w/:workspaceId/p/:projectId/overview', element: previewRoute(<ProjectOverview />) },
  { path: '/w/:workspaceId/p/:projectId/timeline', element: previewRoute(<ProjectTimeline />) },
  { path: '/w/:workspaceId/p/:projectId/calendar', element: withLayout(<ProjectCalendar />) },
  {
    path: '/w/:workspaceId/p/:projectId/dashboard',
    element: previewRoute(<ProjectDashboard />),
  },
  { path: '/w/:workspaceId/p/:projectId/settings', element: withLayout(<ProjectSettings />) },
  { path: '/task/:id', element: withLayout(<TaskDetail />) },
  { path: '/my-issues', element: withLayout(<MyIssues />) },
  { path: '/today', element: withLayout(<FocusToday />) },
  { path: '/timeline', element: previewRoute(<Timeline />) },
  { path: '/calendar', element: withLayout(<CalendarView />) },
  { path: '/milestones', element: withLayout(<Milestones />) },
  { path: '/cycles', element: withLayout(<Cycles />) },
  { path: '/views', element: withLayout(<Views />) },
  { path: '/roadmap', element: previewRoute(<Roadmap />) },
  { path: '/insights', element: previewRoute(<Insights />) },
  { path: '/team', element: previewRoute(<Team />) },
  { path: '/triage', element: withLayout(<Triage />) },
  { path: '/archive', element: withLayout(<Archive />) },
  { path: '/inbox', element: withLayout(<Inbox />) },
  { path: '/settings', element: withLayout(<FlowSettings />) },
  {
    path: '*',
    element: (
      <Navigate to="/w/$workspaceId" params={{ workspaceId: PERSONAL_WORKSPACE_ID }} replace />
    ),
  },
];

export default appRoutes;
