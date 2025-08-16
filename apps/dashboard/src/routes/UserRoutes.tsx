import DashboardLayout from "../layouts/Client/UserDashbaord";
import React from "react";
import { AnimatedRoutes } from "@voxire/shared";
import SettingsPage from "../screens/Client/Settings";

// Routes
const Profile = React.lazy(() => import("../screens/Client/Profile"));
const Cart = React.lazy(() => import("../screens/Client/Cart"));
const Chat = React.lazy(() => import("../screens/Chat/Chat"));

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
  {
    path: "/chat",
    element: (
      <AnimatedRoutes>
        <Chat />
      </AnimatedRoutes>
    ),
  },
];

export default UserRoutes;
