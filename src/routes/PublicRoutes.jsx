import AnimatedRoutes from "../components/AnimatedRoutes";
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
        <AnimatedRoutes>
          <Home />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/about",
    element: (
      <Public>
        <AnimatedRoutes>
          <About />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/pricing",
    element: (
      <Public>
        <AnimatedRoutes>
          <Pricing />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/contact",
    element: (
      <Public>
        <AnimatedRoutes>
          <Contact />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/feedback",
    element: (
      <Public>
        <AnimatedRoutes>
          <Feedback />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
