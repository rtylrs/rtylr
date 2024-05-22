import React from "react";
import Addons from "../screens/Vendor/Addons";
import VendorSettings from "../screens/Vendor/VendorSettings";
import Vendor from "../layouts/Vendor/Vendor";

// Routes
const VendorDashboard = React.lazy(() =>
  import("../screens/Vendor/Dashboard/VendorDashboard")
);

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
];

export default VendorRoutes;
