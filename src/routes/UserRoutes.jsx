import DashboardLayout from "../layouts/Client/UserDashbaord";
import React from "react";
import SettingsPage from "../screens/Client/Settings";

// Routes
const Profile = React.lazy(() => import("../screens/Client/Profile"));
const BecomeBusiness = React.lazy(() =>
  import("../screens/Client/BecomeBusiness")
);

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
    path: "/join",
    element: (
      <DashboardLayout>
        <BecomeBusiness />
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
];

export default UserRoutes;
