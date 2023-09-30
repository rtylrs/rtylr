import React from "react";
import VendorDashboard from "../layouts/Vendor/VendorDashbaord";

// Routes
const VendorMain = React.lazy(() => import("../screens/Vendor/VendorMain"));

const VendorRoutes = [
  {
    path: "/business",
    element: (
      <VendorDashboard>
        <VendorMain />
      </VendorDashboard>
    ),
    exact: true,
  },
];

export default VendorRoutes;
