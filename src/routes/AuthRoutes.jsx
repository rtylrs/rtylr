import React from "react";
import Public from "../layouts/Public/public";
import AnimatedRoutes from "../components/AnimatedRoutes";

// Routes
const Login = React.lazy(() => import("../screens/Auth/Login"));
const Register = React.lazy(() => import("../screens/Auth/Register"));
const ForgotPassword = React.lazy(() => import("../screens/Auth/Reset"));

const AuthRoutes = [
  {
    path: "/auth/login",
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
    path: "/auth/register",
    element: (
      <Public>
        <AnimatedRoutes>
          <Register />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
  {
    path: "/auth/password-reset",
    element: (
      <Public>
        <AnimatedRoutes>
          <ForgotPassword />
        </AnimatedRoutes>
      </Public>
    ),
    exact: true,
  },
];

export default AuthRoutes;
