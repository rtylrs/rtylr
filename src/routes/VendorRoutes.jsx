import React from "react";
import VendorDashboard from "../layouts/Vendor/VendorDashbaord";
import Addons from "../screens/Vendor/Addons";
import VendorSettings from "../screens/Vendor/VendorSettings";

// Routes
const VendorMain = React.lazy(() => import("../screens/Vendor/VendorMain"));

const VendorRoutes = [
  {
    path: "/manager/restaurant/manage/",
    element: (
      <VendorDashboard>
        <VendorMain />
      </VendorDashboard>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/manage/addons",
    element: (
      <VendorDashboard>
        <Addons />
      </VendorDashboard>
    ),
    exact: true,
  },
  {
    path: "/manager/restaurant/manage/settings",
    element: (
      <VendorDashboard>
        <VendorSettings />
      </VendorDashboard>
    ),
    exact: true,
  },
];

export default VendorRoutes;
