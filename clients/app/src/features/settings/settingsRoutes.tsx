import React, { Suspense } from 'react';
import {
  createRoute,
  Navigate,
  Outlet,
  type AnyRoute,
  type ParsedLocation,
} from '@tanstack/react-router';
import { Spinner } from '@voxire/components';
import { AppFlagGuard, FLAG_KEYS, type FlagKey } from '@voxire/config';
import { settingsSearchSchema } from '../../modules/flow/shared/schemas/settings';
import { createHrSettingsRoutes } from '../../modules/hr/settingsRoutes';
import { requireDashManagement } from '../../shared/utils/requireDashManagement';

const SettingsOverview = React.lazy(() => import('./SettingsOverview'));
const FlowSettings = React.lazy(() => import('../../modules/flow/features/settings/FlowSettings'));
const RecruitSettings = React.lazy(
  () => import('../../modules/recruit/features/admin/CustomizationAdmin')
);
const FinanceSettings = React.lazy(
  () => import('../../modules/finance/features/settings/Settings')
);
const ErpSettings = React.lazy(() => import('../../modules/erp/features/settings/Settings'));
const CrmSettings = React.lazy(() => import('../../modules/crm/features/settings/Settings'));
const InsightsSettings = React.lazy(
  () => import('../../modules/insights/features/settings/Settings')
);
const Account = React.lazy(() => import('./Settings/components/Account'));
const Businesses = React.lazy(() => import('./Settings/components/Businesses'));
const General = React.lazy(() => import('./Settings/components/General'));
const Industry = React.lazy(() => import('./Settings/components/BusinessType'));
const Domain = React.lazy(() => import('./Settings/components/Domain/Domain'));
const Email = React.lazy(() => import('./Settings/components/Email/Email'));
const AiUsage = React.lazy(() => import('./Settings/components/AiUsage'));
const Billing = React.lazy(() => import('../billing/Billing.tsx'));
const Security = React.lazy(() => import('../security/Security'));
const ApiKeys = React.lazy(() => import('../developer/ApiKeys'));
const Webhooks = React.lazy(() => import('../developer/Webhooks'));

const fallback = (
  <div className="flex min-h-48 items-center justify-center">
    <Spinner size={32} />
  </div>
);
const wrap = (children: React.ReactNode) => <Suspense fallback={fallback}>{children}</Suspense>;

type SettingsBeforeLoad = (context: { location: ParsedLocation; preload: boolean }) => void;

/**
 * Settings and the account pages listed beside it. The app router mounts these
 * with `openInSettingsModal` so visiting them opens the settings modal; the
 * modal's own router mounts them without it and renders them.
 */
export function createSettingsRoutes<TParent extends AnyRoute>(
  parent: TParent,
  beforeLoad?: SettingsBeforeLoad
) {
  const settingsRoute = createRoute({
    getParentRoute: () => parent,
    path: '/settings',
    beforeLoad,
    component: () => wrap(<Outlet />),
  });

  const settingsChild = <TPath extends string>(path: TPath, element: React.ReactNode) =>
    createRoute({
      getParentRoute: () => settingsRoute,
      path,
      component: () => wrap(element),
    });
  const appSettings = <TPath extends string>(
    path: TPath,
    flag: FlagKey,
    element: React.ReactNode
  ) => settingsChild(path, <AppFlagGuard flag={flag}>{element}</AppFlagGuard>);
  const redirectTo = <TPath extends string>(path: TPath, to: '/settings/billing' | '/extensions') =>
    createRoute({
      getParentRoute: () => settingsRoute,
      path,
      component: () => <Navigate to={to} replace />,
    });

  const settingsTree = settingsRoute.addChildren([
    settingsChild('/', <SettingsOverview />),
    createRoute({
      getParentRoute: () => settingsRoute,
      path: '/flow',
      validateSearch: settingsSearchSchema,
      component: () => wrap(<FlowSettings />),
    }),
    appSettings('/insights', FLAG_KEYS.APP_INSIGHTS, <InsightsSettings />),
    appSettings('/recruit', FLAG_KEYS.APP_RECRUIT, <RecruitSettings />),
    appSettings('/finance', FLAG_KEYS.APP_FINANCE, <FinanceSettings />),
    appSettings('/erp', FLAG_KEYS.APP_ERP, <ErpSettings />),
    appSettings('/crm', FLAG_KEYS.APP_CRM, <CrmSettings />),
    settingsChild('/account', <Account />),
    settingsChild('/businesses', <Businesses />),
    settingsChild('/general', <General />),
    settingsChild('/industry', <Industry />),
    settingsChild('/domain', <Domain />),
    settingsChild('/email', <Email />),
    settingsChild('/ai', <AiUsage />),
    settingsChild('/billing', <Billing />),
    redirectTo('/subscription', '/settings/billing'),
    redirectTo('/store', '/extensions'),
    redirectTo('/addons', '/extensions'),
    redirectTo('/integrations', '/extensions'),
    createHrSettingsRoutes(settingsRoute),
  ]);

  const securityRoute = createRoute({
    getParentRoute: () => parent,
    path: '/security',
    beforeLoad,
    component: () => wrap(<Security />),
  });
  const managedRoute = <TPath extends '/api-keys' | '/webhooks'>(
    path: TPath,
    element: React.ReactNode
  ) =>
    createRoute({
      getParentRoute: () => parent,
      path,
      beforeLoad: (context) => {
        beforeLoad?.(context);
        requireDashManagement();
      },
      component: () => wrap(element),
    });

  return [
    settingsTree,
    securityRoute,
    managedRoute('/api-keys', <ApiKeys />),
    managedRoute('/webhooks', <Webhooks />),
  ] as const;
}
