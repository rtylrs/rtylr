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

const VendorRoutes = [
  {
    path: "/manager/restaurant/",
    element: (
      <Vendor>
        <VendorDashboard />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/addons",
    element: (
      <Vendor>
        <Addons />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/settings",
    element: (
      <Vendor>
        <VendorSettings />
      </Vendor>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/menu",
    element: (
      <Vendor>
        <Menu />
      </Vendor>
    ),
    exact: true,
  },
];

export default VendorRoutes;
