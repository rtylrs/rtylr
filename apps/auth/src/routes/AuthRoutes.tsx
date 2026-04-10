import React from 'react';
import Public from '../layouts/Public/public';
import { AnimatedRoutes } from '@voxire/components';

const Login = React.lazy(() => import('../screens/Login'));
const Register = React.lazy(() => import('../screens/Register'));
const ResetPassword = React.lazy(() => import('../screens/ResetPassword'));
const EmailVerification = React.lazy(
  () => import('../screens/EmailVerification')
);
const VerifyEmail = React.lazy(() => import('../screens/VerifyEmail'));

const wrap = (node: React.ReactNode) => (
  <Public hasHeader={false} hasPadding={false}>
    <AnimatedRoutes>{node}</AnimatedRoutes>
  </Public>
);

const AuthRoutes = [
  { path: '/', element: wrap(<Login />) },
  { path: '/register', element: wrap(<Register />) },
  { path: '/password-reset', element: wrap(<ResetPassword />) },
  { path: '/verify/:token', element: wrap(<EmailVerification />) },
  { path: '/verify', element: wrap(<VerifyEmail />) },
];

export default AuthRoutes;
