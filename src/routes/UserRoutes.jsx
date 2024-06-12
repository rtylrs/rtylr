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
    exact: true,
  },
  {
    path: "/settings",
    element: (
      <DashboardLayout>
        <SettingsPage />
      </DashboardLayout>
    ),
    exact: true,
  },
  {
    path: "/account/cart",
    element: (
      <DashboardLayout>
        <Cart />
      </DashboardLayout>
    ),
    exact: true,
  },
];

export default UserRoutes;
