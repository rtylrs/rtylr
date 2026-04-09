import React from 'react';
import Public from '../layouts/Public/public';
import { AnimatedRoutes } from '@voxire/components';

const Login = React.lazy(() => import('../screens/Login'));
const Register = React.lazy(() => import('../screens/Register'));
const ResetPassword = React.lazy(() => import('../screens/ResetPassword'));
const EmailVerification = React.lazy(
  () => import('../screens/EmailVerification')
);

const AuthRoutes = [
  {
    path: '/',
    element: (
      <Public hasHeader={false} hasPadding={false}>
        <AnimatedRoutes>
          <Login />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: '/register',
    element: (
      <Public hasHeader={false} hasPadding={false}>
        <AnimatedRoutes>
          <Register />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: '/password-reset',
    element: (
      <Public hasHeader={false} hasPadding={false}>
        <AnimatedRoutes>
          <ResetPassword />
        </AnimatedRoutes>
      </Public>
    ),
  },
  {
    path: '/verify/:token',
    element: (
      <Public hasHeader={false} hasPadding={false}>
        <AnimatedRoutes>
          <EmailVerification />
        </AnimatedRoutes>
      </Public>
    ),
  },
];

export default AuthRoutes;
