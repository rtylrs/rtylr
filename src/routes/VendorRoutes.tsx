import React from "react";
import Vendor from "../layouts/Vendor/Vendor";
import MenuWithId from "../screens/Vendor/Menu/MenuWithId";

// Routes
const VendorDashboard = React.lazy(
  () => import("../screens/Vendor/Dashboard/VendorDashboard")
);
const Menu = React.lazy(() => import("../screens/Vendor/Menu/Menu.tsx"));
const Menus = React.lazy(() => import("../screens/Vendor/Menu/Menus"));
const CreateItem = React.lazy(
  () => import("../screens/Vendor/Menu/CreateItem")
);
const CreateBanner = React.lazy(
  () => import("../screens/Vendor/Menu/CreateBanner")
);
const VendorSettings = React.lazy(
  () => import("../screens/Vendor/Settings/VendorSettings")
);
const Addons = React.lazy(() => import("../screens/Vendor/Addons"));
const QrCodes = React.lazy(() => import("../screens/Vendor/QrCode/QrCodes"));
const CreateQrCode = React.lazy(
  () => import("../screens/Vendor/QrCode/CreateQrCode")
);
const Billing = React.lazy(() => import("../screens/Vendor/Billing/Billing"));
const Design = React.lazy(() => import("../screens/Vendor/Design/Design"));

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
        <CreateQrCode />
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
];

export default VendorRoutes;
