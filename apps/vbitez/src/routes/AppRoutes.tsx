import React from 'react';

const POS = React.lazy(() => import('../screens/POS'));
const Payment = React.lazy(() => import('../screens/Payment'));
const Receipt = React.lazy(() => import('../screens/Receipt'));

const AppRoutes = [
  { path: '/', element: <POS /> },
  { path: '/payment', element: <Payment /> },
  { path: '/receipt', element: <Receipt /> },
];

export default AppRoutes;
