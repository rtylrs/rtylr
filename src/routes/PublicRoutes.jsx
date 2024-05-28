import AnimatedRoutes from "../components/AnimatedRoutes";
import Public from "../layouts/Public/public";
import React from "react";
import Collection from "../screens/Public/Collection";

// Routes
const Home = React.lazy(() => import("../screens/Public/Home/Home"));
const About = React.lazy(() => import("../screens/Public/About"));
const Pricing = React.lazy(() => import("../screens/Public/Pricing"));
const Contact = React.lazy(() => import("../screens/Public/Contact"));
const Feedback = React.lazy(() => import("../screens/Public/Feedback"));
const AllBlogs = React.lazy(() => import("../screens/Public/AllBlogs"));
const Faq = React.lazy(() => import("../screens/Public/FAQ"));
const News = React.lazy(() => import("../screens/Public/News"));

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
  {
    path: "/collection/:collectionName",
    element: (
      <Public>
        <Collection />
      </Public>
    ),
    exact: true,
  },
  {
    path: "/all-blogs",
    element: (
      <Public>
        <AnimatedRoutes>
          <AllBlogs />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/faqs",
    element: (
      <Public>
        <AnimatedRoutes>
          <Faq />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/news/:newsId",
    element: (
      <Public>
        <AnimatedRoutes>
          <News />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
