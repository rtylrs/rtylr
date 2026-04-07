import React from "react";
import AppLayout from "../layouts/AppLayout";
import MenuWithId from "../screens/Vendor/Menu/MenuWithId.jsx";

// Routes
const VendorDashboard = React.lazy(
  () => import("../screens/Vendor/Dashboard/VendorDashboard.tsx")
);
const Menu = React.lazy(() => import("../screens/Vendor/Menu/Menu.tsx"));
const Menus = React.lazy(() => import("../screens/Vendor/Menu/Menus.tsx"));
const CreateItem = React.lazy(
  () => import("../screens/Vendor/Menu/Components/CreateItem.jsx")
);
const CreateBanner = React.lazy(
  () => import("../screens/Vendor/Menu/Components/CreateBanner.tsx")
);
const VendorSettings = React.lazy(
  () => import("../screens/Vendor/Settings/VendorSettings.tsx")
);
const Addons = React.lazy(() => import("../screens/Vendor/Addons.tsx"));
const QrCodes = React.lazy(
  () => import("../screens/Vendor/QrCode/QrCodes.tsx")
);
const QrCode = React.lazy(() => import("../screens/Vendor/QrCode/QrCode.tsx"));
const Billing = React.lazy(
  () => import("../screens/Vendor/Billing/Billing.tsx")
);
const Design = React.lazy(() => import("../screens/Vendor/Design/Design.tsx"));
const General = React.lazy(
  () => import("../screens/Vendor/Settings/General.tsx")
);

const AppRoutes = [
  {
    path: "/",
    element: (
      <AppLayout>
        <VendorDashboard />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/addons",
    element: (
      <AppLayout>
        <Addons />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/settings",
    element: (
      <AppLayout>
        <VendorSettings />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/menus",
    element: (
      <AppLayout>
        <Menus />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/menu",
    element: (
      <AppLayout>
        <Menu />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/menu/:id",
    element: (
      <AppLayout>
        <MenuWithId />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/create-item",
    element: (
      <AppLayout>
        <CreateItem />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/create-banner",
    element: (
      <AppLayout>
        <CreateBanner />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/qr-codes",
    element: (
      <AppLayout>
        <QrCodes />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/qr-codes/create",
    element: (
      <AppLayout>
        <QrCode />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/billing",
    element: (
      <AppLayout>
        <Billing />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/design",
    element: (
      <AppLayout>
        <Design />
      </AppLayout>
    ),
  },
  {
    path: "/app/business/settings/general",
    element: (
      <AppLayout>
        <General />
      </AppLayout>
    ),
  },
];

export default AppRoutes;
