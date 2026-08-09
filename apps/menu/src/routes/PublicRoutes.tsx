import React from 'react';

const PosMenu = React.lazy(() => import('../screens/Public/PosMenu/PosMenu'));
const SlugLanding = React.lazy(() => import('../screens/Public/SlugLanding/SlugLanding'));
const LocationMenu = React.lazy(() => import('../screens/Public/SlugLanding/LocationMenu'));
const NoMatch = React.lazy(() => import('../screens/Public/Extras/NoMatch'));
const DevMenu = React.lazy(() => import('../screens/Dev/DevMenu'));
const Checkout = React.lazy(() => import('../features/cart/Checkout'));
const OrderStatus = React.lazy(() => import('../features/cart/OrderStatus'));

const isLocalhost =
  typeof window !== 'undefined' &&
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1');

const PublicRoutes = [
  // POS-backed public menu - numeric business id under /b/ so the
  // path is unambiguous vs the slug-or-legacy-id catch-all.
  { path: '/b/:businessId', element: <PosMenu /> },

  // Online ordering flow - accessible from any menu
  { path: '/checkout', element: <Checkout /> },
  { path: '/order-status/:id', element: <OrderStatus /> },

  // New public surface: menu.rtylr.com/<business-slug>/<location-slug>
  { path: '/:slug/:location', element: <LocationMenu /> },

  // menu.rtylr.com/<slug> - landing that lists active locations.
  // SlugLanding also handles the legacy 24-char mongo ID fallback so
  // pre-existing QR codes keep resolving on the same surface.
  { path: '/:slug', element: <SlugLanding /> },

  // Dev preview, only available on localhost
  ...(isLocalhost ? [{ path: '/', element: <DevMenu /> }] : [{ path: '/', element: <NoMatch /> }]),

  { path: '*', element: <NoMatch /> },
];

export default PublicRoutes;
