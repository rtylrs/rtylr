import React, { Suspense } from 'react';
import { z } from 'zod';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  Navigate,
  redirect,
} from '@tanstack/react-router';
import { ErrorBoundary } from '@voxire/ui';
import { Spinner } from '@voxire/components';
import { BusinessRoute } from './shared/utils/BusinessRoute';
import { canManageDash } from './shared/utils/canManageDash';
import { useUserStore } from './shared/stores/userStore';
import SettingsLayout from './features/settings/Settings/components/SettingsLayout';
import NoMatch from './features/extras/NoMatch';
import { createInsightsRoutes } from './modules/insights/routes';

const Orders = React.lazy(() => import('./features/orders/Orders'));
const Home = React.lazy(() => import('./features/home/Home'));
const Menu = React.lazy(() => import('./features/menu/Menu.tsx'));
const Menus = React.lazy(() => import('./features/menu/Menus.tsx'));
const CreateItem = React.lazy(() => import('./features/menu/Components/CreateItem.jsx'));
const CreateBanner = React.lazy(() => import('./features/menu/Components/CreateBanner.tsx'));
const QrCodes = React.lazy(() => import('./features/qr-codes/QrCodes.tsx'));
const QrCode = React.lazy(() => import('./features/qr-codes/QrCode.tsx'));
const Design = React.lazy(() => import('./features/design/Design.tsx'));
const Chat = React.lazy(() => import('./features/chat/Chat'));

const Account = React.lazy(() => import('./features/settings/Settings/components/Account'));
const Businesses = React.lazy(() => import('./features/settings/Settings/components/Businesses'));
const General = React.lazy(() => import('./features/settings/Settings/components/General'));
const Industry = React.lazy(() => import('./features/settings/Settings/components/BusinessType'));
const Domain = React.lazy(() => import('./features/settings/Settings/components/Domain/Domain'));
const Email = React.lazy(() => import('./features/settings/Settings/components/Email/Email'));
const AiUsage = React.lazy(() => import('./features/settings/Settings/components/AiUsage'));
const Billing = React.lazy(() => import('./features/billing/Billing.tsx'));
const Addons = React.lazy(() => import('./features/addons/Addons.tsx'));

const Team = React.lazy(() => import('./features/team/Team'));
const InviteUser = React.lazy(() => import('./features/team/InviteUser'));
const MemberDetail = React.lazy(() => import('./features/team/MemberDetail'));
const Groups = React.lazy(() => import('./features/team/Groups'));
const GroupDetail = React.lazy(() => import('./features/team/GroupDetail'));
const Roles = React.lazy(() => import('./features/permissions/Roles'));
const Audit = React.lazy(() => import('./features/activity/Audit'));
const Notifications = React.lazy(() => import('./features/notifications/Notifications'));
const Integrations = React.lazy(() => import('./features/integrations/Integrations'));
const ApiKeys = React.lazy(() => import('./features/developer/ApiKeys'));
const Webhooks = React.lazy(() => import('./features/developer/Webhooks'));
const Security = React.lazy(() => import('./features/security/Security'));

const ThemeCustomizer = React.lazy(() => import('./features/extras/Theme'));
const HelpCenter = React.lazy(() => import('./features/extras/HelpCenter'));

const fallback = (
  <div className="flex min-h-48 items-center justify-center">
    <Spinner size={32} />
  </div>
);

const wrap = (children: React.ReactNode) => <Suspense fallback={fallback}>{children}</Suspense>;

const requireDashManagement = () => {
  const user = useUserStore.getState().user;
  // Deny until Dash access has been hydrated; BusinessRoute handles the session.
  if (!canManageDash(user)) {
    throw redirect({ to: '/' });
  }
};

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  ),
});

const businessRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'business',
  component: BusinessRoute,
});

const dashboardRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/',
  component: () => wrap(<Home />),
});

const menusRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/menus',
  component: () => wrap(<Menus />),
});

const menuRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/menu',
  component: () => wrap(<Menu />),
});

const createItemRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/create-item',
  component: () => wrap(<CreateItem />),
});

const createBannerRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/create-banner',
  component: () => wrap(<CreateBanner />),
});

const ordersRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/orders',
  component: () => wrap(<Orders />),
});

const qrCodesRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/qr-codes',
  component: () => wrap(<QrCodes />),
});

const qrCodeCreateRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/qr-codes/create',
  component: () => wrap(<QrCode />),
});

const designRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/design',
  component: () => wrap(<Design />),
});

const chatRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/chat',
  component: () => wrap(<Chat />),
});

const teamRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team',
  component: () => wrap(<Team />),
});

const teamInviteRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team/invite',
  component: () => wrap(<InviteUser />),
});

const teamGroupsRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team/groups',
  component: () => wrap(<Groups />),
});

const teamGroupNewRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team/groups/new',
  component: () => wrap(<GroupDetail />),
});

const teamGroupDetailRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team/groups/$id',
  component: () => wrap(<GroupDetail />),
});

const teamMemberDetailRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/team/$id',
  component: () => wrap(<MemberDetail />),
});

const rolesRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/roles',
  validateSearch: z.object({
    create: z.literal(1).optional().catch(undefined),
  }),
  beforeLoad: requireDashManagement,
  component: () => wrap(<Roles />),
});

const auditRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/audit',
  beforeLoad: requireDashManagement,
  component: () => wrap(<Audit />),
});

const notificationsRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/notifications',
  component: () => wrap(<Notifications />),
});

const accountRedirectRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/account',
  component: () => <Navigate to="/settings/account" replace />,
});

const billingRedirectRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/billing',
  component: () => <Navigate to="/settings/billing" replace />,
});

const apiKeysRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/api-keys',
  beforeLoad: requireDashManagement,
  component: () => wrap(<ApiKeys />),
});

const webhooksRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/webhooks',
  beforeLoad: requireDashManagement,
  component: () => wrap(<Webhooks />),
});

const securityRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/security',
  component: () => wrap(<Security />),
});

const themeRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/theme',
  component: () => wrap(<ThemeCustomizer />),
});

const helpRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/help',
  component: () => wrap(<HelpCenter />),
});

const integrationsRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/integrations',
  component: () => wrap(<Integrations />),
});

// Settings layout routes
const settingsRoute = createRoute({
  getParentRoute: () => businessRoute,
  path: '/settings',
  component: () => (
    <Suspense fallback={fallback}>
      <SettingsLayout />
    </Suspense>
  ),
});

const settingsIndexRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/',
  component: () => <Navigate to="/settings/account" replace />,
});

const settingsAccountRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/account',
  component: () => (
    <Suspense fallback={fallback}>
      <Account />
    </Suspense>
  ),
});

const settingsBusinessesRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/businesses',
  component: () => (
    <Suspense fallback={fallback}>
      <Businesses />
    </Suspense>
  ),
});

const settingsGeneralRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/general',
  component: () => (
    <Suspense fallback={fallback}>
      <General />
    </Suspense>
  ),
});

const settingsIndustryRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/industry',
  component: () => (
    <Suspense fallback={fallback}>
      <Industry />
    </Suspense>
  ),
});

const settingsDomainRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/domain',
  component: () => (
    <Suspense fallback={fallback}>
      <Domain />
    </Suspense>
  ),
});

const settingsEmailRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/email',
  component: () => (
    <Suspense fallback={fallback}>
      <Email />
    </Suspense>
  ),
});

const settingsAiRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/ai',
  component: () => (
    <Suspense fallback={fallback}>
      <AiUsage />
    </Suspense>
  ),
});

const settingsSubscriptionRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/subscription',
  component: () => <Navigate to="/settings/billing" replace />,
});

const settingsBillingRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/billing',
  component: () => (
    <Suspense fallback={fallback}>
      <Billing />
    </Suspense>
  ),
});

const settingsAddonsRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/addons',
  component: () => (
    <Suspense fallback={fallback}>
      <Addons />
    </Suspense>
  ),
});

const settingsIntegrationsRoute = createRoute({
  getParentRoute: () => settingsRoute,
  path: '/integrations',
  component: () => (
    <Suspense fallback={fallback}>
      <Integrations />
    </Suspense>
  ),
});

const notFoundRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '*',
  component: NoMatch,
});

const routeTree = rootRoute.addChildren([
  businessRoute.addChildren([
    createInsightsRoutes(businessRoute),
    dashboardRoute,
    menusRoute,
    menuRoute,
    createItemRoute,
    createBannerRoute,
    ordersRoute,
    qrCodesRoute,
    qrCodeCreateRoute,
    designRoute,
    chatRoute,
    teamRoute,
    teamInviteRoute,
    teamGroupsRoute,
    teamGroupNewRoute,
    teamGroupDetailRoute,
    teamMemberDetailRoute,
    rolesRoute,
    auditRoute,
    notificationsRoute,
    accountRedirectRoute,
    billingRedirectRoute,
    apiKeysRoute,
    webhooksRoute,
    securityRoute,
    themeRoute,
    helpRoute,
    integrationsRoute,
    settingsRoute.addChildren([
      settingsIndexRoute,
      settingsAccountRoute,
      settingsBusinessesRoute,
      settingsGeneralRoute,
      settingsIndustryRoute,
      settingsDomainRoute,
      settingsEmailRoute,
      settingsAiRoute,
      settingsSubscriptionRoute,
      settingsBillingRoute,
      settingsAddonsRoute,
      settingsIntegrationsRoute,
    ]),
  ]),
  notFoundRoute,
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
