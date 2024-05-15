import React from "react";
import VendorDashboard from "../layouts/Vendor/VendorDashbaord";
import Addons from "../screens/Vendor/Addons";
import VendorSettings from "../screens/Vendor/VendorSettings";

// Routes
const VendorMain = React.lazy(() => import("../screens/Vendor/VendorMain"));

const VendorRoutes = [
  {
    path: "/manager/restaurant/",
    element: (
      <VendorDashboard>
        <VendorMain />
      </VendorDashboard>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/addons",
    element: (
      <VendorDashboard>
        <Addons />
      </VendorDashboard>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/settings",
    element: (
      <VendorDashboard>
        <VendorSettings />
      </VendorDashboard>
    ),
    exact: true,
  },
];

export default VendorRoutes;
