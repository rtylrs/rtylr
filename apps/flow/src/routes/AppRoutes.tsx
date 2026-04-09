import React, { Suspense } from 'react';
import { RouteObject, Navigate } from 'react-router-dom';
import FlowLayout from '../layouts/FlowLayout/FlowLayout';

const WorkspaceDashboard = React.lazy(
  () => import('../screens/WorkspaceDashboard/WorkspaceDashboard')
);
const ProjectBoard = React.lazy(
  () => import('../screens/ProjectBoard/ProjectBoard')
);
const ProjectList = React.lazy(
  () => import('../screens/ProjectList/ProjectList')
);
const ProjectOverview = React.lazy(
  () => import('../screens/ProjectOverview/ProjectOverview')
);
const ProjectSettings = React.lazy(
  () => import('../screens/ProjectSettings/ProjectSettings')
);
const Inbox = React.lazy(() => import('../screens/Inbox/Inbox'));
const FlowSettings = React.lazy(
  () => import('../screens/FlowSettings/FlowSettings')
);
const TaskDetail = React.lazy(() => import('../screens/TaskDetail/TaskDetail'));
const MyIssues = React.lazy(() => import('../screens/MyIssues/MyIssues'));
const Timeline = React.lazy(() => import('../screens/Timeline/Timeline'));
const CalendarView = React.lazy(() => import('../screens/Calendar/Calendar'));
const Milestones = React.lazy(() => import('../screens/Milestones/Milestones'));
const Cycles = React.lazy(() => import('../screens/Cycles/Cycles'));
const Views = React.lazy(() => import('../screens/Views/Views'));
const Roadmap = React.lazy(() => import('../screens/Roadmap/Roadmap'));
const Insights = React.lazy(() => import('../screens/Insights/Insights'));
const Team = React.lazy(() => import('../screens/Team/Team'));
const Triage = React.lazy(() => import('../screens/Triage/Triage'));
const Archive = React.lazy(() => import('../screens/Archive/Archive'));

function withLayout(element: React.ReactNode): React.ReactNode {
  return (
    <FlowLayout>
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
    </FlowLayout>
  );
}

const appRoutes: RouteObject[] = [
  { path: '/', element: <Navigate to="/w/default" replace /> },
  { path: '/w/:workspaceId', element: withLayout(<WorkspaceDashboard />) },
  { path: '/w/:workspaceId/p/:projectId', element: withLayout(<ProjectBoard />) },
  { path: '/w/:workspaceId/p/:projectId/list', element: withLayout(<ProjectList />) },
  { path: '/w/:workspaceId/p/:projectId/overview', element: withLayout(<ProjectOverview />) },
  { path: '/w/:workspaceId/p/:projectId/settings', element: withLayout(<ProjectSettings />) },
  { path: '/task/:id', element: withLayout(<TaskDetail />) },
  { path: '/my-issues', element: withLayout(<MyIssues />) },
  { path: '/timeline', element: withLayout(<Timeline />) },
  { path: '/calendar', element: withLayout(<CalendarView />) },
  { path: '/milestones', element: withLayout(<Milestones />) },
  { path: '/cycles', element: withLayout(<Cycles />) },
  { path: '/views', element: withLayout(<Views />) },
  { path: '/roadmap', element: withLayout(<Roadmap />) },
  { path: '/insights', element: withLayout(<Insights />) },
  { path: '/team', element: withLayout(<Team />) },
  { path: '/triage', element: withLayout(<Triage />) },
  { path: '/archive', element: withLayout(<Archive />) },
  { path: '/inbox', element: withLayout(<Inbox />) },
  { path: '/settings', element: withLayout(<FlowSettings />) },
  { path: '*', element: <Navigate to="/w/default" replace /> },
];

export default appRoutes;
