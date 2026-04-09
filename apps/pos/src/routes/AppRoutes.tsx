import React from 'react';
import POSLayout from '../layouts/POSLayout/POSLayout';
import Terminal from '../screens/Terminal/Terminal';
import Orders from '../screens/Orders/Orders';
import Tables from '../screens/Tables/Tables';
import CloseSession from '../screens/CloseSession/CloseSession';

interface RouteConfig {
  path: string;
  element: React.ReactNode;
}

const AppRoutes: RouteConfig[] = [
  {
    path: '/',
    element: (
      <POSLayout>
        <Terminal />
      </POSLayout>
    ),
  },
  {
    path: '/orders',
    element: (
      <POSLayout>
        <Orders />
      </POSLayout>
    ),
  },
  {
    path: '/tables',
    element: (
      <POSLayout>
        <Tables />
      </POSLayout>
    ),
  },
  {
    path: '/close',
    element: (
      <POSLayout>
        <CloseSession />
      </POSLayout>
    ),
  },
];

export default AppRoutes;
