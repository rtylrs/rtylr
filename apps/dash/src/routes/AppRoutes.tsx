import React from 'react';
import { RouteObject, Navigate } from 'react-router-dom';
import AppLayout from '../layouts/AppLayout';
import MenuWithId from '../features/menu/MenuWithId.jsx';
import SettingsLayout from '../features/settings/Settings/components/SettingsLayout';

const Orders = React.lazy(() => import('../features/orders/Orders'));
const VendorDashboard = React.lazy(() => import('../features/dashboard/Dashboard/VendorDashboard'));
const Menu = React.lazy(() => import('../features/menu/Menu.tsx'));
const Menus = React.lazy(() => import('../features/menu/Menus.tsx'));
const CreateItem = React.lazy(() => import('../features/menu/Components/CreateItem.jsx'));
const CreateBanner = React.lazy(() => import('../features/menu/Components/CreateBanner.tsx'));
const QrCodes = React.lazy(() => import('../features/qr-codes/QrCodes.tsx'));
const QrCode = React.lazy(() => import('../features/qr-codes/QrCode.tsx'));
const Design = React.lazy(() => import('../features/design/Design.tsx'));
const Analytics = React.lazy(() => import('../features/analytics/Analytics.tsx'));
const Chat = React.lazy(() => import('../features/chat/Chat'));

// Settings children
const Account = React.lazy(() => import('../features/settings/Settings/components/Account'));
const General = React.lazy(() => import('../features/settings/Settings/components/General'));
const Industry = React.lazy(() => import('../features/settings/Settings/components/BusinessType'));
const Domain = React.lazy(() => import('../features/settings/Settings/components/Domain/Domain'));
const Email = React.lazy(() => import('../features/settings/Settings/components/Email/Email'));
const Billing = React.lazy(() => import('../features/billing/Billing.tsx'));
const Addons = React.lazy(() => import('../features/addons/Addons.tsx'));
const Subscription = React.lazy(() => import('../features/subscription/Subscription'));

// New Hub screens
const Team = React.lazy(() => import('../features/team/Team'));
const InviteUser = React.lazy(() => import('../features/team/InviteUser'));
const MemberDetail = React.lazy(() => import('../features/team/MemberDetail'));
const Permissions = React.lazy(() => import('../features/permissions/Permissions'));
const Roles = React.lazy(() => import('../features/permissions/Roles'));
const AppsScreen = React.lazy(() => import('../features/apps/Apps'));
const ActivityScreen = React.lazy(() => import('../features/activity/Activity'));
const Audit = React.lazy(() => import('../features/activity/Audit'));
const Notifications = React.lazy(() => import('../features/notifications/Notifications'));
const AccountHome = React.lazy(() => import('../features/account/Account'));
const BillingHome = React.lazy(() => import('../features/billing/BillingHome'));
const Integrations = React.lazy(() => import('../features/integrations/Integrations'));
const ApiKeys = React.lazy(() => import('../features/developer/ApiKeys'));
const Webhooks = React.lazy(() => import('../features/developer/Webhooks'));
const Security = React.lazy(() => import('../features/security/Security'));

// Extras
const Marketplace = React.lazy(() => import('../features/extras/Marketplace'));
const Changelog = React.lazy(() => import('../features/extras/Changelog'));
const FeatureFlags = React.lazy(() => import('../features/extras/FeatureFlags'));
const ThemeCustomizer = React.lazy(() => import('../features/extras/Theme'));
const Referrals = React.lazy(() => import('../features/extras/Referrals'));
const HelpCenter = React.lazy(() => import('../features/extras/HelpCenter'));
const DataExport = React.lazy(() =>
  import('../features/extras/DataPorts').then((m) => ({ default: m.default }))
);
const DataImport = React.lazy(() =>
  import('../features/extras/DataPorts').then((m) => ({ default: m.DataImport }))
);

const wrap = (children: React.ReactNode) => <AppLayout>{children}</AppLayout>;

const AppRoutes: RouteObject[] = [
  { path: '/', element: wrap(<VendorDashboard />) },
  { path: '/menus', element: wrap(<Menus />) },
  { path: '/menu', element: wrap(<Menu />) },
  { path: '/menu/:id', element: wrap(<MenuWithId />) },
  { path: '/create-item', element: wrap(<CreateItem />) },
  { path: '/create-banner', element: wrap(<CreateBanner />) },
  { path: '/orders', element: wrap(<Orders />) },
  { path: '/qr-codes', element: wrap(<QrCodes />) },
  { path: '/qr-codes/create', element: wrap(<QrCode />) },
  { path: '/design', element: wrap(<Design />) },
  { path: '/analytics', element: wrap(<Analytics />) },
  { path: '/chat', element: wrap(<Chat />) },

  // New Hub screens
  { path: '/team', element: wrap(<Team />) },
  { path: '/team/invite', element: wrap(<InviteUser />) },
  { path: '/team/:id', element: wrap(<MemberDetail />) },
  { path: '/permissions', element: wrap(<Permissions />) },
  { path: '/roles', element: wrap(<Roles />) },
  { path: '/apps', element: wrap(<AppsScreen />) },
  { path: '/activity', element: wrap(<ActivityScreen />) },
  { path: '/audit', element: wrap(<Audit />) },
  { path: '/notifications', element: wrap(<Notifications />) },
  { path: '/account', element: <Navigate to="/settings/account" replace /> },
  { path: '/billing', element: <Navigate to="/settings/billing" replace /> },
  { path: '/api-keys', element: wrap(<ApiKeys />) },
  { path: '/webhooks', element: wrap(<Webhooks />) },
  { path: '/security', element: wrap(<Security />) },

  // Extras
  { path: '/theme', element: wrap(<ThemeCustomizer />) },
  { path: '/help', element: wrap(<HelpCenter />) },
  {
    path: '/settings',
    element: <SettingsLayout />,
    children: [
      { index: true, element: <Navigate to="/settings/account" replace /> },
      { path: 'account', element: <Account /> },
      { path: 'general', element: <General /> },
      { path: 'industry', element: <Industry /> },
      { path: 'domain', element: <Domain /> },
      { path: 'email', element: <Email /> },
      { path: 'subscription', element: <Subscription /> },
      { path: 'billing', element: <Billing /> },
      { path: 'addons', element: <Addons /> },
      { path: 'integrations', element: <Integrations /> },
    ],
  },
];

export default AppRoutes;
