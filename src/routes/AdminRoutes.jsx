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

const Users = React.lazy(() => import("../screens/Admin/AdminUsers"));
const AdminContacts = React.lazy(() =>
  import("../screens/Admin/AdminContacts")
);
const AdminSettings = React.lazy(() =>
  import("../screens/Admin/AdminSettings")
);

const AboutConfig = React.lazy(() =>
  import("../screens/Admin/config/AboutConfig")
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
  {
    path: "/manage/users",
    element: (
      <Admin>
        <Users />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/contacts",
    element: (
      <Admin>
        <AdminContacts />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/settings",
    element: (
      <Admin>
        <AdminSettings />
      </Admin>
    ),
    exact: true,
  },
  {
    path: "/manage/config/about",
    element: (
      <Admin>
        <AboutConfig />
      </Admin>
    ),
    exact: true,
  },
];

export default AdminRoutes;
