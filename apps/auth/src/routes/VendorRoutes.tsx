import React from "react";
import Vendor from "../layouts/Vendor/Vendor.tsx";
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
const Billing = React.lazy(() => import("../screens/Vendor/Billing/Billing.tsx"));
const Design = React.lazy(() => import("../screens/Vendor/Design/Design.tsx"));
const General = React.lazy(() => import("../screens/Vendor/Settings/General.tsx"));

const VendorRoutes = [
  {
    path: "/app/business/",
    element: (
      <Vendor>
        <VendorDashboard />
      </Vendor>
    ),
  },
  {
    path: "/app/business/addons",
    element: (
      <Vendor>
        <Addons />
      </Vendor>
    ),
  },
  {
    path: "/app/business/settings",
    element: (
      <Vendor>
        <VendorSettings />
      </Vendor>
    ),
  },
  {
    path: "/app/business/menus",
    element: (
      <Vendor>
        <Menus />
      </Vendor>
    ),
  },
  {
    path: "/app/business/menu",
    element: (
      <Vendor>
        <Menu />
      </Vendor>
    ),
  },
  {
    path: "/app/business/menu/:id",
    element: (
      <Vendor>
        <MenuWithId />
      </Vendor>
    ),
  },
  {
    path: "/app/business/create-item",
    element: (
      <Vendor>
        <CreateItem />
      </Vendor>
    ),
  },
  {
    path: "/app/business/create-banner",
    element: (
      <Vendor>
        <CreateBanner />
      </Vendor>
    ),
  },
  {
    path: "/app/business/qr-codes",
    element: (
      <Vendor>
        <QrCodes />
      </Vendor>
    ),
  },
  {
    path: "/app/business/qr-codes/create",
    element: (
      <Vendor>
        <QrCode />
      </Vendor>
    ),
  },
  {
    path: "/app/business/billing",
    element: (
      <Vendor>
        <Billing />
      </Vendor>
    ),
  },
  {
    path: "/app/business/design",
    element: (
      <Vendor>
        <Design />
      </Vendor>
    ),
  },
  {
    path: "/app/business/settings/general",
    element: (
      <Vendor>
        <General />
      </Vendor>
    ),
  },
];

export default VendorRoutes;
