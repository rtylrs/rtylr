import React from "react";
import Vendor from "../layouts/Vendor/Vendor";
import MenuWithId from "../screens/Vendor/Menu/MenuWithId";

// Routes
const VendorDashboard = React.lazy(() =>
  import("../screens/Vendor/Dashboard/VendorDashboard")
);
const Menu = React.lazy(() => import("../screens/Vendor/Menu/Menu"));
const Menus = React.lazy(() => import("../screens/Vendor/Menu/Menus"));
const CreateItem = React.lazy(() =>
  import("../screens/Vendor/Menu/CreateItem")
);
const VendorSettings = React.lazy(() =>
  import("../screens/Vendor/Settings/VendorSettings")
);
const Addons = React.lazy(() => import("../screens/Vendor/Addons"));
const QrCodes = React.lazy(() => import("../screens/Vendor/QrCode/QrCodes"));
const CreateQrCode = React.lazy(() =>
  import("../screens/Vendor/QrCode/CreateQrCode")
);
const Billing = React.lazy(() => import("../screens/Vendor/Billing/Billing"));
const Design = React.lazy(() => import("../screens/Vendor/Design/Design"));

const VendorRoutes = [
  {
    path: "/app/restaurant/",
    element: (
      <Vendor>
        <VendorDashboard />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/addons",
    element: (
      <Vendor>
        <Addons />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/settings",
    element: (
      <Vendor>
        <VendorSettings />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/menus",
    element: (
      <Vendor>
        <Menus />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/menu",
    element: (
      <Vendor>
        <Menu />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/menu/:id",
    element: (
      <Vendor>
        <MenuWithId />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/create-item",
    element: (
      <Vendor>
        <CreateItem />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/qr-codes",
    element: (
      <Vendor>
        <QrCodes />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/qr-codes/create",
    element: (
      <Vendor>
        <CreateQrCode />
      </Vendor>
    ),
  },
  {
    path: "/app/restaurant/billing",
    element: (
      <Vendor>
        <Billing />
      </Vendor>
    ),
  },

  {
    path: "/app/restaurant/design",
    element: (
      <Vendor>
        <Design />
      </Vendor>
    ),
  },
];

export default VendorRoutes;
