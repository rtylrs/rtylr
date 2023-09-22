import Public from "../layouts/Public/public";
import React from "react";

// Routes
const Home = React.lazy(() => import("../screens/Public/Home"));
const About = React.lazy(() => import("../screens/Public/About"));
const Pricing = React.lazy(() => import("../screens/Public/Pricing"));
const Contact = React.lazy(() => import("../screens/Public/Contact"));
const Feedback = React.lazy(() => import("../screens/Public/Feedback"));

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
  {
    path: "/pricing",
    element: (
      <Public>
        <Pricing />
      </Public>
    ),
    exact: true,
  },
  {
    path: "/contact",
    element: (
      <Public>
        <Contact />
      </Public>
    ),
    exact: true,
  },
  {
    path: "/feedback",
    element: (
      <Public>
        <Feedback />
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
