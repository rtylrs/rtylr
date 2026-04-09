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
const Inbox = React.lazy(() => import('../screens/Inbox/Inbox'));
const FlowSettings = React.lazy(
  () => import('../screens/FlowSettings/FlowSettings')
);

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
  {
    path: '/',
    element: <Navigate to="/w/default" replace />,
  },
  {
    path: '/w/:workspaceId',
    element: withLayout(<WorkspaceDashboard />),
  },
  {
    path: '/w/:workspaceId/p/:projectId',
    element: withLayout(<ProjectBoard />),
  },
  {
    path: '/w/:workspaceId/p/:projectId/list',
    element: withLayout(<ProjectList />),
  },
  {
    path: '/inbox',
    element: withLayout(<Inbox />),
  },
  {
    path: '/settings',
    element: withLayout(<FlowSettings />),
  },
  {
    path: '*',
    element: <Navigate to="/w/default" replace />,
  },
];

export default appRoutes;
