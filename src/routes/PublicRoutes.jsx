import Public from "../layouts/Public/public";
import React from "react";

// Routes
const Home = React.lazy(() => import("../screens/Public/Home"));

const PublicRoutes = [
  {
    path: "/",
    element: (
      <Public>
        <Home />
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
