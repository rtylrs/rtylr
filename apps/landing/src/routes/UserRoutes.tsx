import DashboardLayout from "../layouts/Client/UserDashbaord";
import React from "react";
import SettingsPage from "../screens/Client/Settings";

// Routes
const Profile = React.lazy(() => import("../screens/Client/Profile"));
const Cart = React.lazy(() => import("../screens/Client/Cart"));

const UserRoutes = [
  {
    path: "/account",
    element: (
      <DashboardLayout>
        <Profile />
      </DashboardLayout>
    ),
  },
  {
    path: "/settings",
    element: (
      <DashboardLayout>
        <SettingsPage />
      </DashboardLayout>
    ),
  },
  {
    path: "/account/cart",
    element: (
      <DashboardLayout>
        <Cart />
      </DashboardLayout>
    ),
  },
];

export default UserRoutes;
