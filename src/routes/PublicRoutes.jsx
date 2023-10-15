import AnimatedRoutes from "../components/AnimatedRoutes";
import Public from "../layouts/Public/public";
import React from "react";
import Collection from "../screens/Public/Collection";

// Routes
const Home = React.lazy(() => import("../screens/Public/Home"));
const Shop = React.lazy(() => import("../screens/Public/Shop/Shop"));
const ShopItemPage = React.lazy(() =>
  import("../screens/Public/Shop/ShopItemPage")
);
const About = React.lazy(() => import("../screens/Public/About"));
const Pricing = React.lazy(() => import("../screens/Public/Pricing"));
const Contact = React.lazy(() => import("../screens/Public/Contact"));
const Feedback = React.lazy(() => import("../screens/Public/Feedback"));
const AllPetitions = React.lazy(() => import("../screens/Public/AllPetitions"));
const AllBlogs = React.lazy(() => import("../screens/Public/AllBlogs"));
const Events = React.lazy(() => import("../screens/Public/Events"));

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
    path: "/shop",
    element: (
      <Public>
        <AnimatedRoutes>
          <Shop />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/shop/:itemId",
    element: (
      <Public>
        <AnimatedRoutes>
          <ShopItemPage />
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
    path: "/all-petitions",
    element: (
      <Public>
        <AnimatedRoutes>
          <AllPetitions />
        </AnimatedRoutes>
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
    path: "/events",
    element: (
      <Public>
        <AnimatedRoutes>
          <Events />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
];

export default PublicRoutes;
