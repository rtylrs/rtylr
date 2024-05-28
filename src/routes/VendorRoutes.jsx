import React from "react";
import Vendor from "../layouts/Vendor/Vendor";

// Routes
const VendorDashboard = React.lazy(() =>
  import("../screens/Vendor/Dashboard/VendorDashboard")
);
const Menu = React.lazy(() => import("../screens/Vendor/Menu/Menu"));
const VendorSettings = React.lazy(() =>
  import("../screens/Vendor/VendorSettings")
);
const Addons = React.lazy(() => import("../screens/Vendor/Addons"));
const QrCodes = React.lazy(() => import("../screens/Vendor/QrCode/QrCodes"));
const CreateQrCode = React.lazy(() =>
  import("../screens/Vendor/QrCode/CreateQrCode")
);

const VendorRoutes = [
  {
    path: "/app/restaurant/",
    element: (
      <Vendor>
        <VendorDashboard />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/app/restaurant/addons",
    element: (
      <Vendor>
        <Addons />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/app/restaurant/settings",
    element: (
      <Vendor>
        <VendorSettings />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/app/restaurant/menu",
    element: (
      <Vendor>
        <Menu />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/app/restaurant/qr-codes",
    element: (
      <Vendor>
        <QrCodes />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/app/restaurant/qr-codes/create",
    element: (
      <Vendor>
        <CreateQrCode />
      </Vendor>
    ),
    exact: true,
  },
];

export default VendorRoutes;
