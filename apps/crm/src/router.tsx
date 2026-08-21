import React, { Suspense } from 'react';
import {
  createRootRoute,
  createRoute,
  createRouter,
  Navigate,
  Outlet,
} from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { AppServiceClient } from '@voxire/voxie/hub/app/v1/client';
import type { Addon } from '@voxire/voxie/hub/app/v1/types';
import { FullScreenLoader } from '@voxire/components';
import { z } from 'zod';
import { ShortcutsProvider } from './shared/components/ShortcutsProvider';
import { userClient } from './shared/lib/clients';
import DefaultLayout from './layouts/DefaultLayout';

const Dashboard = React.lazy(() => import('./features/dashboard/Dashboard'));
const Customers = React.lazy(() => import('./features/customers/Customers'));
const CustomerProfile = React.lazy(() => import('./features/customers/CustomerProfile'));
const CreateCustomer = React.lazy(() => import('./features/customers/CreateCustomer'));
const ImportCustomers = React.lazy(() => import('./features/customers/ImportCustomers'));
const MergeCustomers = React.lazy(() => import('./features/customers/MergeCustomers'));
const Loyalty = React.lazy(() => import('./features/loyalty/Loyalty'));
const LoyaltyMembers = React.lazy(() => import('./features/loyalty/LoyaltyMembers'));
const LoyaltyRewards = React.lazy(() => import('./features/loyalty/LoyaltyRewards'));
const LoyaltyTiers = React.lazy(() => import('./features/loyalty/LoyaltyTiers'));
const Campaigns = React.lazy(() => import('./features/campaigns/Campaigns'));
const EmailCampaign = React.lazy(() => import('./features/campaigns/EmailCampaign'));
const SMSCampaign = React.lazy(() => import('./features/campaigns/SMSCampaign'));
const CampaignDetail = React.lazy(() => import('./features/campaigns/CampaignDetail'));
const Segments = React.lazy(() => import('./features/segments/Segments'));
const SegmentBuilder = React.lazy(() => import('./features/segments/SegmentBuilder'));
const SegmentDetail = React.lazy(() => import('./features/segments/SegmentDetail'));
const Automations = React.lazy(() => import('./features/automations/Automations'));
const AutomationBuilder = React.lazy(() => import('./features/automations/AutomationBuilder'));
const Templates = React.lazy(() => import('./features/templates/Templates'));
const Feedback = React.lazy(() => import('./features/feedback/Feedback'));
const SettingsScreen = React.lazy(() => import('./features/settings/Settings'));
const Pipeline = React.lazy(() => import('./features/sales/Pipeline'));
const Leads = React.lazy(() => import('./features/sales/Leads'));
const LeadDetail = React.lazy(() => import('./features/sales/LeadDetail'));
const SalesContracts = React.lazy(() => import('./features/sales/Contracts'));

function lazy(element: React.ReactNode): React.ReactNode {
  return <Suspense fallback={<FullScreenLoader />}>{element}</Suspense>;
}

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <AppGuard
        appId="crm"
        fetchUser={async (): Promise<Record<string, unknown>> =>
          (await userClient.getMe({})) as unknown as Record<string, unknown>
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_CRM}>
          <ShortcutsProvider>
            <Outlet />
          </ShortcutsProvider>
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  ),
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'layout',
  component: DefaultLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: () => <>{lazy(<Dashboard />)}</>,
});

// Customers
const customersRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers',
  component: () => <>{lazy(<Customers />)}</>,
});

const customersNewRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers/new',
  component: () => <>{lazy(<CreateCustomer />)}</>,
});

const customersImportRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers/import',
  component: () => <>{lazy(<ImportCustomers />)}</>,
});

const customersMergeRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers/merge',
  component: () => <>{lazy(<MergeCustomers />)}</>,
});

const customerProfileRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/customers/$id',
  component: () => <>{lazy(<CustomerProfile />)}</>,
});

// Loyalty
const loyaltyRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/loyalty',
  component: () => <>{lazy(<Loyalty />)}</>,
});

const loyaltyMembersRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/loyalty/members',
  component: () => <>{lazy(<LoyaltyMembers />)}</>,
});

const loyaltyRewardsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/loyalty/rewards',
  component: () => <>{lazy(<LoyaltyRewards />)}</>,
});

const loyaltyTiersRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/loyalty/tiers',
  component: () => <>{lazy(<LoyaltyTiers />)}</>,
});

// Campaigns
const campaignsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/campaigns',
  component: () => <>{lazy(<Campaigns />)}</>,
});

const campaignsEmailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/campaigns/email',
  component: () => <>{lazy(<EmailCampaign />)}</>,
});

const campaignsSmsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/campaigns/sms',
  component: () => <>{lazy(<SMSCampaign />)}</>,
});

const campaignDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/campaigns/$id',
  component: () => <>{lazy(<CampaignDetail />)}</>,
});

// Segments
const segmentsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/segments',
  component: () => <>{lazy(<Segments />)}</>,
});

const segmentsNewRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/segments/new',
  component: () => <>{lazy(<SegmentBuilder />)}</>,
});

const segmentDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/segments/$id',
  component: () => <>{lazy(<SegmentDetail />)}</>,
});

// Automations
const automationsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/automations',
  component: () => <>{lazy(<Automations />)}</>,
});

const automationsNewRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/automations/new',
  validateSearch: z.object({
    id: z.number().int().positive().optional().catch(undefined),
  }),
  component: () => <>{lazy(<AutomationBuilder />)}</>,
});

// Templates, feedback, settings
const templatesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/templates',
  component: () => <>{lazy(<Templates />)}</>,
});

const feedbackRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/feedback',
  component: () => <>{lazy(<Feedback />)}</>,
});

const settingsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/settings',
  component: () => <>{lazy(<SettingsScreen />)}</>,
});

const appClient = createOneKitClient(AppServiceClient, 'crm');

function SalesAddonGuard({ children }: { children: React.ReactNode }): React.JSX.Element {
  const { data, isLoading } = useQuery({
    queryKey: ['crm-addons'],
    queryFn: async () => {
      try {
        const res = await appClient.listAddons({});
        return res.addons ?? [];
      } catch {
        return [];
      }
    },
    staleTime: 5 * 60 * 1000,
  });

  if (isLoading) return <FullScreenLoader />;

  const isEnabled = (data ?? []).some((a: Addon) => a.id === 'sales_pipeline' && a.isEnabled);
  if (!isEnabled) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}

function guardSales(component: React.ReactNode): React.JSX.Element {
  return <SalesAddonGuard>{lazy(component)}</SalesAddonGuard>;
}

// Sales Pipeline
const salesRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/sales',
  component: () => guardSales(<Pipeline />),
});

const salesLeadsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/sales/leads',
  component: () => guardSales(<Leads />),
});

const salesLeadDetailRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/sales/leads/$id',
  component: () => guardSales(<LeadDetail />),
});

const salesContractsRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/sales/contracts',
  component: () => guardSales(<SalesContracts />),
});

const catchAllRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '*',
  component: () => <Navigate to="/" replace />,
});

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([
    indexRoute,
    customersRoute,
    customersNewRoute,
    customersImportRoute,
    customersMergeRoute,
    customerProfileRoute,
    loyaltyRoute,
    loyaltyMembersRoute,
    loyaltyRewardsRoute,
    loyaltyTiersRoute,
    campaignsRoute,
    campaignsEmailRoute,
    campaignsSmsRoute,
    campaignDetailRoute,
    segmentsRoute,
    segmentsNewRoute,
    segmentDetailRoute,
    automationsRoute,
    automationsNewRoute,
    templatesRoute,
    feedbackRoute,
    settingsRoute,
    salesRoute,
    salesLeadsRoute,
    salesLeadDetailRoute,
    salesContractsRoute,
    catchAllRoute,
  ]),
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
