import React from 'react';

const Menu = React.lazy(() => import('../screens/Public/Menu/Menu'));
const PosMenu = React.lazy(() => import('../screens/Public/PosMenu/PosMenu'));
const NoMatch = React.lazy(() => import('../screens/Public/Extras/NoMatch'));
const DevMenu = React.lazy(() => import('../screens/Dev/DevMenu'));
const Checkout = React.lazy(() => import('../features/cart/Checkout'));
const OrderStatus = React.lazy(() => import('../features/cart/OrderStatus'));

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

const PublicRoutes = [
  // POS-backed public menu — numeric business id under /b/ so the
  // path is unambiguous vs the legacy mongo-id catch-all `:id`.
  {
    path: '/b/:businessId',
    element: <PosMenu />,
  },
  // Online ordering flow — accessible from any menu
  {
    path: '/checkout',
    element: <Checkout />,
  },
  {
    path: '/order-status/:id',
    element: <OrderStatus />,
  },
  // Legacy mongo-backed public menu. Kept intact so existing QR
  // codes continue to resolve. New businesses land on /b/{id}.
  {
    path: '/:id',
    element: <Menu />,
  },
  // Dev preview, only available on localhost
  ...(isLocalhost ? [{ path: '/', element: <DevMenu /> }] : [{ path: '/', element: <NoMatch /> }]),
  {
    path: '*',
    element: <NoMatch />,
  },
];

export default PublicRoutes;
