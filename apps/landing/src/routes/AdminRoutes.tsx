import React from "react";
import Admin from "../layouts/Admin/Admin";

// Routes
const Dashboard = React.lazy(() => import("../screens/Admin/Dashboard"));
const AdminBlog = React.lazy(() => import("../screens/Admin/AdminBlog"));
const AdminFeedback = React.lazy(
  () => import("../screens/Admin/AdminFeedBack")
);
const AdminProducts = React.lazy(
  () => import("../screens/Admin/AdminProducts")
);

const Users = React.lazy(() => import("../screens/Admin/AdminUsers"));

const AdminSettings = React.lazy(
  () => import("../screens/Admin/AdminSettings")
);

const AboutConfig = React.lazy(
  () => import("../screens/Admin/config/AboutConfig")
);

const AddFaq = React.lazy(() => import("../screens/Admin/FAQs/AddFaq"));
const Faqs = React.lazy(() => import("../screens/Admin/FAQs/Faqs"));

const AdminRoutes = [
  {
    path: "/manage",
    element: (
      <Admin>
        <Dashboard />
      </Admin>
    ),
  },
  {
    path: "/manage/blog",
    element: (
      <Admin>
        <AdminBlog />
      </Admin>
    ),
  },
  {
    path: "/manage/feedback",
    element: (
      <Admin>
        <AdminFeedback />
      </Admin>
    ),
  },
  {
    path: "/manage/products",
    element: (
      <Admin>
        <AdminProducts />
      </Admin>
    ),
  },
  {
    path: "/manage/users",
    element: (
      <Admin>
        <Users />
      </Admin>
    ),
  },
  {
    path: "/manage/settings",
    element: (
      <Admin>
        <AdminSettings />
      </Admin>
    ),
  },
  {
    path: "/manage/config/about",
    element: (
      <Admin>
        <AboutConfig />
      </Admin>
    ),
  },
  {
    path: "/manage/sections/faqs",
    element: (
      <Admin>
        <Faqs />
      </Admin>
    ),
  },
  {
    path: "/manage/sections/faqs/add",
    element: (
      <Admin>
        <AddFaq />
      </Admin>
    ),
  },
];

export default AdminRoutes;
