import React from 'react';
import Public from '../layouts/Public/public';
import { AnimatedRoutes } from '@voxire/components';

const Welcome = React.lazy(() => import('../screens/Welcome'));
const Login = React.lazy(() => import('../screens/Login'));
const Register = React.lazy(() => import('../screens/Register'));
const ResetPassword = React.lazy(() => import('../screens/ResetPassword'));
const EmailVerification = React.lazy(
  () => import('../screens/EmailVerification')
);
const TwoFactorSetup = React.lazy(() => import('../screens/TwoFactorSetup'));
const TwoFactorVerify = React.lazy(() => import('../screens/TwoFactorVerify'));
const MagicLink = React.lazy(() => import('../screens/MagicLink'));
const MagicLinkConfirm = React.lazy(
  () => import('../screens/MagicLinkConfirm')
);
const SSO = React.lazy(() => import('../screens/SSO'));
const ForgotUsername = React.lazy(() => import('../screens/ForgotUsername'));
const ChangeEmail = React.lazy(() => import('../screens/ChangeEmail'));
const Sessions = React.lazy(() => import('../screens/Sessions'));
const TrustedDevices = React.lazy(() => import('../screens/TrustedDevices'));
const SecurityQuestions = React.lazy(
  () => import('../screens/SecurityQuestions')
);
const AccountRecovery = React.lazy(() => import('../screens/AccountRecovery'));
const InvitationAccept = React.lazy(
  () => import('../screens/InvitationAccept')
);
const VerifyEmail = React.lazy(() => import('../screens/VerifyEmail'));
const SignupComplete = React.lazy(() => import('../screens/SignupComplete'));
const MobileHandoff = React.lazy(() => import('../screens/MobileHandoff'));

const wrap = (node: React.ReactNode) => (
  <Public hasHeader={false} hasPadding={false}>
    <AnimatedRoutes>{node}</AnimatedRoutes>
  </Public>
);

const AuthRoutes = [
  { path: '/', element: wrap(<Welcome />) },
  { path: '/login', element: wrap(<Login />) },
  { path: '/register', element: wrap(<Register />) },
  { path: '/password-reset', element: wrap(<ResetPassword />) },
  { path: '/verify/:token', element: wrap(<EmailVerification />) },

  { path: '/2fa/setup', element: wrap(<TwoFactorSetup />) },
  { path: '/2fa/verify', element: wrap(<TwoFactorVerify />) },

  { path: '/magic-link', element: wrap(<MagicLink />) },
  { path: '/magic-link/confirm', element: wrap(<MagicLinkConfirm />) },

  { path: '/sso', element: wrap(<SSO />) },
  { path: '/forgot-username', element: wrap(<ForgotUsername />) },
  { path: '/change-email', element: wrap(<ChangeEmail />) },
  { path: '/sessions', element: wrap(<Sessions />) },
  { path: '/devices', element: wrap(<TrustedDevices />) },
  { path: '/security-questions', element: wrap(<SecurityQuestions />) },
  { path: '/recovery', element: wrap(<AccountRecovery />) },
  { path: '/invite/:token', element: wrap(<InvitationAccept />) },
  { path: '/verify', element: wrap(<VerifyEmail />) },
  { path: '/welcome', element: wrap(<SignupComplete />) },
  { path: '/mobile-handoff', element: wrap(<MobileHandoff />) },
];

export default AuthRoutes;
