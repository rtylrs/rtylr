import React from "react";
import { RouteObject, Navigate } from "react-router-dom";
import AppLayout from "../layouts/AppLayout";
import MenuWithId from "../screens/Vendor/Menu/MenuWithId.jsx";
import SettingsLayout from "../screens/Vendor/Settings/components/SettingsLayout";

const Orders        = React.lazy(() => import("../screens/Vendor/Orders/Orders"));
const VendorDashboard = React.lazy(() => import("../screens/Vendor/Dashboard/VendorDashboard.tsx"));
const Menu          = React.lazy(() => import("../screens/Vendor/Menu/Menu.tsx"));
const Menus         = React.lazy(() => import("../screens/Vendor/Menu/Menus.tsx"));
const CreateItem    = React.lazy(() => import("../screens/Vendor/Menu/Components/CreateItem.jsx"));
const CreateBanner  = React.lazy(() => import("../screens/Vendor/Menu/Components/CreateBanner.tsx"));
const QrCodes       = React.lazy(() => import("../screens/Vendor/QrCode/QrCodes.tsx"));
const QrCode        = React.lazy(() => import("../screens/Vendor/QrCode/QrCode.tsx"));
const Design        = React.lazy(() => import("../screens/Vendor/Design/Design.tsx"));
const Analytics     = React.lazy(() => import("../screens/Vendor/Analytics/Analytics.tsx"));
const Chat          = React.lazy(() => import("../screens/Vendor/Chat/Chat"));

// Settings children
const Account       = React.lazy(() => import("../screens/Vendor/Settings/components/Account"));
const General       = React.lazy(() => import("../screens/Vendor/Settings/components/General"));
const Industry      = React.lazy(() => import("../screens/Vendor/Settings/components/BusinessType"));
const Domain        = React.lazy(() => import("../screens/Vendor/Settings/components/Domain/Domain"));
const Email         = React.lazy(() => import("../screens/Vendor/Settings/components/Email/Email"));
const Billing       = React.lazy(() => import("../screens/Vendor/Billing/Billing.tsx"));
const Addons        = React.lazy(() => import("../screens/Vendor/Addons.tsx"));

// New Hub screens
const Team              = React.lazy(() => import("../screens/Vendor/Team/Team"));
const InviteUser        = React.lazy(() => import("../screens/Vendor/Team/InviteUser"));
const MemberDetail      = React.lazy(() => import("../screens/Vendor/Team/MemberDetail"));
const Permissions       = React.lazy(() => import("../screens/Vendor/Permissions/Permissions"));
const Roles             = React.lazy(() => import("../screens/Vendor/Permissions/Roles"));
const AppsScreen        = React.lazy(() => import("../screens/Vendor/Apps/Apps"));
const ActivityScreen    = React.lazy(() => import("../screens/Vendor/Activity/Activity"));
const Audit             = React.lazy(() => import("../screens/Vendor/Activity/Audit"));
const Notifications     = React.lazy(() => import("../screens/Vendor/Notifications/Notifications"));
const AccountHome       = React.lazy(() => import("../screens/Vendor/Account/Account"));
const BillingHome       = React.lazy(() => import("../screens/Vendor/Billing/BillingHome"));
const Integrations      = React.lazy(() => import("../screens/Vendor/Integrations/Integrations"));
const ApiKeys           = React.lazy(() => import("../screens/Vendor/Developer/ApiKeys"));
const Webhooks          = React.lazy(() => import("../screens/Vendor/Developer/Webhooks"));
const Security          = React.lazy(() => import("../screens/Vendor/Security/Security"));

// Extras
const Marketplace       = React.lazy(() => import("../screens/Vendor/Extras/Marketplace"));
const Changelog         = React.lazy(() => import("../screens/Vendor/Extras/Changelog"));
const FeatureFlags      = React.lazy(() => import("../screens/Vendor/Extras/FeatureFlags"));
const ThemeCustomizer   = React.lazy(() => import("../screens/Vendor/Extras/Theme"));
const Referrals         = React.lazy(() => import("../screens/Vendor/Extras/Referrals"));
const HelpCenter        = React.lazy(() => import("../screens/Vendor/Extras/HelpCenter"));
const DataExport        = React.lazy(() => import("../screens/Vendor/Extras/DataPorts").then((m) => ({ default: m.default })));
const DataImport        = React.lazy(() => import("../screens/Vendor/Extras/DataPorts").then((m) => ({ default: m.DataImport })));

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
  { path: '/chat',      element: wrap(<Chat />) },

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
      { path: 'billing', element: <Billing /> },
      { path: 'addons', element: <Addons /> },
      { path: 'integrations', element: <Integrations /> },
    ],
  },
];

export default AppRoutes;
