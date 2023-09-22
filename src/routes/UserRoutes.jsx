import DashboardLayout from "../layouts/Client/UserDashbaord";

// Routes
import Profile from "../screens/Client/Profile";

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
];

export default UserRoutes;
