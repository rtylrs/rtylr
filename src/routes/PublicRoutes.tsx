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
const Menu = React.lazy(() => import("../screens/Public/Menu/Menu"));
const UseCases = React.lazy(() => import("../screens/Public/UseCases"));

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
  },
  {
    path: "/collection/:collectionName",
    element: (
      <Public>
        <Collection />
      </Public>
    ),
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
  },
  {
    path: "/restaurant/:id",
    element: (
      <AnimatedRoutes>
        <Menu />
      </AnimatedRoutes>
    ),
  },
  {
    path: "/uses",
    element: (
      <Public>
        <AnimatedRoutes>
          <UseCases />
        </AnimatedRoutes>
      </Public>
    ),
  },
];

export default PublicRoutes;
