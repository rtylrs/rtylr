import React from "react";
import Public from "../layouts/Public/public";
import AnimatedRoutes from "../components/AnimatedRoutes";

// Routes
const Login = React.lazy(() => import("../screens/Auth/Login"));
const Register = React.lazy(() => import("../screens/Auth/Register"));

const AuthRoutes = [
  {
    path: "/login",
    element: (
      <Public>
        <AnimatedRoutes>
          <Login />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/register",
    element: (
      <Public>
        <AnimatedRoutes>
          <Register />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
];

export default AuthRoutes;
