import React from 'react';

const Menu = React.lazy(() => import('../screens/Public/Menu/Menu'));
const NoMatch = React.lazy(() => import('../screens/Public/Extras/NoMatch'));
const DevMenu = React.lazy(() => import('../screens/Dev/DevMenu'));

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

const PublicRoutes = [
  {
    path: '/:id',
    element: <Menu />,
  },
  // Dev preview — only available on localhost
  ...(isLocalhost ? [{ path: '/', element: <DevMenu /> }] : [{ path: '/', element: <NoMatch /> }]),
  {
    path: '*',
    element: <NoMatch />,
  },
];

export default PublicRoutes;
