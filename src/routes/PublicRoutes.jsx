import Public from "../layouts/Public/public";
import React from "react";

// Routes
const Home = React.lazy(() => import("../screens/Public/Home"));
const About = React.lazy(() => import("../screens/Public/About"));

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
  {
    path: "/about",
    element: (
      <Public>
        <About />
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
