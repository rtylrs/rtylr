import React from 'react';
import Public from '../layouts/Public/public';
import { AnimatedRoutes } from '@voxire/components';

const Login = React.lazy(() => import('../features/login/Login'));
const Register = React.lazy(() => import('../features/register/Register'));
const ResetPassword = React.lazy(() => import('../features/reset-password/ResetPassword'));
const EmailVerification = React.lazy(
  () => import('../features/email-verification/EmailVerification')
);
const VerifyEmail = React.lazy(() => import('../features/email-verification/VerifyEmail'));
const InviteAccept = React.lazy(() => import('../features/invite/InviteAccept'));

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
  { path: '/invite/:token', element: wrap(<InviteAccept />) },
];

export default AuthRoutes;
