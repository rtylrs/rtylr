import Public from "../layouts/Public/public";

// Routes
import Home from "../screens/Public/Home";

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
