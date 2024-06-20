import React from "react";
import Public from "../layouts/Public/public";
import AnimatedRoutes from "../components/AnimatedRoutes";

// Routes
const Login = React.lazy(() => import("../screens/Auth/Login"));
const RegisterUser = React.lazy(() => import("../screens/Auth/RegisterUser"));
const RegisterRest = React.lazy(() => import("../screens/Auth/RegisterRest"));
const ForgotPassword = React.lazy(() => import("../screens/Auth/Reset"));
const EmailVerification = React.lazy(() =>
  import("../screens/Auth/EmailVerification")
);

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
  },
  {
    path: "/auth/register/user",
    element: (
      <Public>
        <AnimatedRoutes>
          <RegisterUser />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: "/auth/register/restaurant",
    element: (
      <Public>
        <AnimatedRoutes>
          <RegisterRest />
        </AnimatedRoutes>
      </Public>
    ),
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
  },
  {
    path: "/auth/verify/:token",
    element: (
      <Public>
        <AnimatedRoutes>
          <EmailVerification />
        </AnimatedRoutes>
      </Public>
    ),
  },
];

export default AuthRoutes;
