import Admin from "../layouts/Admin/Admin";
import Dashboard from "../screens/Admin/Dashboard";

// Routes

const AdminRoutes = [
  {
    path: "/admin",
    element: (
      <Admin>
        <Dashboard />
      </Admin>
    ),
    exact: true,
  },
];

export default AdminRoutes;
