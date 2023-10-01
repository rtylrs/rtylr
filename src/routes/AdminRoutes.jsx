import React from "react";
import Admin from "../layouts/Admin/Admin";

// Routes
const Dashboard = React.lazy(() => import("../screens/Admin/Dashboard"));
const AdminBlog = React.lazy(() => import("../screens/Admin/AdminBlog"));
const AdminFeedback = React.lazy(() =>
  import("../screens/Admin/AdminFeedBack")
);
const AdminProducts = React.lazy(() =>
  import("../screens/Admin/AdminProducts")
);

const AdminRoutes = [
  {
    path: "/manage",
    element: (
      <Admin>
        <Dashboard />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/blog",
    element: (
      <Admin>
        <AdminBlog />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/feedback",
    element: (
      <Admin>
        <AdminFeedback />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/products",
    element: (
      <Admin>
        <AdminProducts />
      </Admin>
    ),
    exact: true,
  },
];

export default AdminRoutes;
