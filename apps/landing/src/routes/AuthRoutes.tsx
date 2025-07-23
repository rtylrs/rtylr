import React from "react";
import Public from "../layouts/Public/public";
import AnimatedRoutes from "../components/AnimatedRoutes";

// Routes
const Login = React.lazy(() => import("../screens/Auth/Login"));
const Register = React.lazy(() => import("../screens/Auth/Register"));
const ForgotPassword = React.lazy(() => import("../screens/Auth/Reset"));
const EmailVerification = React.lazy(
  () => import("../screens/Auth/EmailVerification")
);
// Questionnaire Route screens
const Details = React.lazy(
  () => import("../screens/Auth/Questionnaire/Details")
);
const QuestionsOne = React.lazy(
  () => import("../screens/Auth/Questionnaire/QuestionsOne")
);
const QuestionsTwo = React.lazy(
  () => import("../screens/Auth/Questionnaire/QuestionsTwo")
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
    path: "/auth/register",
    element: (
      <Public>
        <AnimatedRoutes>
          <Register />
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
  {
    path: "/auth/questionnaire/info",
    element: (
      <Public>
        <AnimatedRoutes>
          <Details />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: "/auth/questionnaire/info/question-1",
    element: (
      <Public>
        <AnimatedRoutes>
          <QuestionsOne />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: "/auth/questionnaire/info/question-2",
    element: (
      <Public>
        <AnimatedRoutes>
          <QuestionsTwo />
        </AnimatedRoutes>
      </Public>
    ),
  },
];

export default AuthRoutes;
