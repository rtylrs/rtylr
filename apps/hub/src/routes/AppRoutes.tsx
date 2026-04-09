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

// Settings children
const Account       = React.lazy(() => import("../screens/Vendor/Settings/components/Account"));
const General       = React.lazy(() => import("../screens/Vendor/Settings/components/General"));
const Industry      = React.lazy(() => import("../screens/Vendor/Settings/components/BusinessType"));
const Domain        = React.lazy(() => import("../screens/Vendor/Settings/components/Domain/Domain"));
const Email         = React.lazy(() => import("../screens/Vendor/Settings/components/Email/Email"));
const Billing       = React.lazy(() => import("../screens/Vendor/Billing/Billing.tsx"));
const Addons        = React.lazy(() => import("../screens/Vendor/Addons.tsx"));

const wrap = (children: React.ReactNode) => <AppLayout>{children}</AppLayout>;

const AppRoutes: RouteObject[] = [
  { path: "/",              element: wrap(<VendorDashboard />) },
  { path: "/menus",         element: wrap(<Menus />) },
  { path: "/menu",          element: wrap(<Menu />) },
  { path: "/menu/:id",      element: wrap(<MenuWithId />) },
  { path: "/create-item",   element: wrap(<CreateItem />) },
  { path: "/create-banner", element: wrap(<CreateBanner />) },
  { path: "/orders",        element: wrap(<Orders />) },
  { path: "/qr-codes",      element: wrap(<QrCodes />) },
  { path: "/qr-codes/create", element: wrap(<QrCode />) },
  { path: "/design",        element: wrap(<Design />) },
  { path: "/analytics",     element: wrap(<Analytics />) },

  // Legacy redirects
  { path: "/billing",  element: <Navigate to="/settings/billing"  replace /> },
  { path: "/addons",   element: <Navigate to="/settings/addons"   replace /> },
  { path: "/app/business/menus",           element: <Navigate to="/menus"          replace /> },
  { path: "/app/business/menu",            element: <Navigate to="/menu"           replace /> },
  { path: "/app/business/menu/:id",        element: <Navigate to="/menu/:id"       replace /> },
  { path: "/app/business/orders",          element: <Navigate to="/orders"         replace /> },
  { path: "/app/business/qr-codes",        element: <Navigate to="/qr-codes"       replace /> },
  { path: "/app/business/qr-codes/create", element: <Navigate to="/qr-codes/create" replace /> },
  { path: "/app/business/billing",         element: <Navigate to="/settings/billing" replace /> },
  { path: "/app/business/design",          element: <Navigate to="/design"         replace /> },
  { path: "/app/business/analytics",       element: <Navigate to="/analytics"      replace /> },
  { path: "/app/business/addons",          element: <Navigate to="/settings/addons" replace /> },
  { path: "/app/business/create-item",     element: <Navigate to="/create-item"    replace /> },
  { path: "/app/business/create-banner",   element: <Navigate to="/create-banner"  replace /> },

  {
    path: "/settings",
    element: <SettingsLayout />,
    children: [
      { index: true,            element: <Navigate to="/settings/account" replace /> },
      { path: "account",        element: <Account /> },
      { path: "general",        element: <General /> },
      { path: "industry",       element: <Industry /> },
      { path: "domain",         element: <Domain /> },
      { path: "email",          element: <Email /> },
      { path: "billing",        element: <Billing /> },
      { path: "addons",         element: <Addons /> },
      // legacy
      { path: "business-type",  element: <Navigate to="/settings/industry" replace /> },
    ],
  },
];

export default AppRoutes;
