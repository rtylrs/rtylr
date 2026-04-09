import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Customers from '../screens/Customers/Customers';
import Loyalty from '../screens/Loyalty/Loyalty';
import Campaigns from '../screens/Campaigns/Campaigns';
import Segments from '../screens/Segments/Segments';

interface RouteChild {
  index?: boolean;
  path?: string;
  element: React.ReactNode;
}

interface AppRoute {
  path: string;
  element: React.ReactNode;
  children?: RouteChild[];
}

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'customers', element: <Customers /> },
      { path: 'loyalty', element: <Loyalty /> },
      { path: 'campaigns', element: <Campaigns /> },
      { path: 'segments', element: <Segments /> },
    ],
  },
];

export default AppRoutes;
