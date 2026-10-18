import React, { Suspense } from 'react';
import { createRoute, Navigate, Outlet, type AnyRoute } from '@tanstack/react-router';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
import { Spinner } from '@voxire/ui';
import { z } from 'zod';
import TrialBanner from './shared/components/TrialBanner';
import { getCrmHomePath } from './navigation';
import { useUserStore } from '../../shared/stores/userStore';

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
const Templates = React.lazy(() => import('./features/templates/Templates'));
const Feedback = React.lazy(() => import('./features/feedback/Feedback'));
const Pipeline = React.lazy(() => import('./features/sales/Pipeline'));
const Leads = React.lazy(() => import('./features/sales/Leads'));
const LeadDetail = React.lazy(() => import('./features/sales/LeadDetail'));
const SalesContracts = React.lazy(() => import('./features/sales/Contracts'));

function lazy(element: React.ReactNode): React.ReactNode {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-48 items-center justify-center">
          <Spinner />
        </div>
      }
    >
      {element}
    </Suspense>
  );
}

function CrmHome(): React.JSX.Element {
  const user = useUserStore((state) => state.user);
  const home = getCrmHomePath(user);
  return home === '/crm' ? <>{lazy(<Dashboard />)}</> : <Navigate to={home} replace />;
}

export function createCrmRoutes<TParent extends AnyRoute>(parent: TParent) {
  const layoutRoute = createRoute({
    getParentRoute: () => parent,
    path: 'crm',
    component: () => (
      <AppFlagGuard flag={FLAG_KEYS.APP_CRM}>
        <ErrorBoundary>
          <TrialBanner />
          <Outlet />
        </ErrorBoundary>
      </AppFlagGuard>
    ),
  });

  const indexRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/',
    component: CrmHome,
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

  // Campaign
  const campaignsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/campaigns',
    component: () => <>{lazy(<Campaigns />)}</>,
  });

  // Editor deep-links carry optional context (?id= for edit mode, ?segment=
  // when arriving from a segment, ?name/?subject/?description/?body= prefills).
  // Validated so stale or hand-edited URLs degrade to defaults instead of erroring.
  const campaignEditorSearchSchema = z.object({
    id: z.number().int().positive().optional().catch(undefined),
    name: z.string().optional().catch(undefined),
    subject: z.string().optional().catch(undefined),
    description: z.string().optional().catch(undefined),
    body: z.string().optional().catch(undefined),
    segment: z.string().optional().catch(undefined),
  });

  const campaignsEmailRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/campaigns/email',
    validateSearch: campaignEditorSearchSchema,
    component: () => <>{lazy(<EmailCampaign />)}</>,
  });

  const campaignsSmsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/campaigns/sms',
    validateSearch: campaignEditorSearchSchema,
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
    component: () => <Navigate to="/automations" replace />,
  });

  const automationsNewRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/automations/new',
    validateSearch: z.object({
      id: z.number().int().positive().optional().catch(undefined),
    }),
    component: () => (
      <Navigate to="/automations/new" search={automationsNewRoute.useSearch()} replace />
    ),
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
    component: () => <Navigate to="/settings/crm" replace />,
  });

  // Sales pipeline. Core CRM, gated by canAccessCrmPath like every other CRM
  // page; it used to also need a `sales_pipeline` add-on the catalog never had.
  const salesRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/sales',
    component: () => lazy(<Pipeline />),
  });

  const salesLeadsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/sales/leads',
    component: () => lazy(<Leads />),
  });

  const salesLeadDetailRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/sales/leads/$id',
    component: () => lazy(<LeadDetail />),
  });

  const salesContractsRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '/sales/contracts',
    component: () => lazy(<SalesContracts />),
  });

  const catchAllRoute = createRoute({
    getParentRoute: () => layoutRoute,
    path: '$',
    component: () => <Navigate to="/crm" replace />,
  });

  return layoutRoute.addChildren([
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
  ]);
}
