import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Invoices from '../screens/Invoices/Invoices';
import Expenses from '../screens/Expenses/Expenses';
import ProfitLoss from '../screens/ProfitLoss/ProfitLoss';
import Tax from '../screens/Tax/Tax';

interface RouteChild {
  index?: boolean;
  path?: string;
  element: React.ReactElement;
}

interface AppRoute {
  path: string;
  element: React.ReactElement;
  children?: RouteChild[];
}

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true,              element: <Dashboard /> },
      { path: 'invoices',        element: <Invoices /> },
      { path: 'expenses',        element: <Expenses /> },
      { path: 'pl',              element: <ProfitLoss /> },
      { path: 'tax',             element: <Tax /> },
    ],
  },
];

export default AppRoutes;
