import React, { Suspense } from 'react';
import { FullScreenLoader } from '@voxire/components';
import DefaultLayout from '../layouts/DefaultLayout';

// All route-level components are lazy-loaded per CLAUDE.md mandate. Layout
// components stay eager (always needed).
const Dashboard = React.lazy(() => import('../features/dashboard/Dashboard'));
const Customers = React.lazy(() => import('../features/customers/Customers'));
const CustomerProfile = React.lazy(() => import('../features/customers/CustomerProfile'));
const CreateCustomer = React.lazy(() => import('../features/customers/CreateCustomer'));
const ImportCustomers = React.lazy(() => import('../features/customers/ImportCustomers'));
const MergeCustomers = React.lazy(() => import('../features/customers/MergeCustomers'));
const Loyalty = React.lazy(() => import('../features/loyalty/Loyalty'));
const LoyaltyMembers = React.lazy(() => import('../features/loyalty/LoyaltyMembers'));
const LoyaltyRewards = React.lazy(() => import('../features/loyalty/LoyaltyRewards'));
const LoyaltyTiers = React.lazy(() => import('../features/loyalty/LoyaltyTiers'));
const Campaigns = React.lazy(() => import('../features/campaigns/Campaigns'));
const EmailCampaign = React.lazy(() => import('../features/campaigns/EmailCampaign'));
const SMSCampaign = React.lazy(() => import('../features/campaigns/SMSCampaign'));
const CampaignDetail = React.lazy(() => import('../features/campaigns/CampaignDetail'));
const Segments = React.lazy(() => import('../features/segments/Segments'));
const SegmentBuilder = React.lazy(() => import('../features/segments/SegmentBuilder'));
const SegmentDetail = React.lazy(() => import('../features/segments/SegmentDetail'));
const Automations = React.lazy(() => import('../features/automations/Automations'));
const AutomationBuilder = React.lazy(() => import('../features/automations/AutomationBuilder'));
const Templates = React.lazy(() => import('../features/templates/Templates'));
const Feedback = React.lazy(() => import('../features/feedback/Feedback'));
const SettingsScreen = React.lazy(() => import('../features/settings/Settings'));

// Sales Pipeline
const Pipeline = React.lazy(() => import('../features/sales/Pipeline'));
const Leads = React.lazy(() => import('../features/sales/Leads'));
const LeadDetail = React.lazy(() => import('../features/sales/LeadDetail'));
const SalesContracts = React.lazy(() => import('../features/sales/Contracts'));

interface RouteChild {
  index?: boolean;
  path?: string;
  element: React.ReactNode;
}

interface AppRoute {
  path: string;
  element: React.ReactNode;
  children?: RouteChild[];
}

function lazy(element: React.ReactNode): React.ReactNode {
  return <Suspense fallback={<FullScreenLoader />}>{element}</Suspense>;
}

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true, element: lazy(<Dashboard />) },

      // Customers
      { path: 'customers', element: lazy(<Customers />) },
      { path: 'customers/new', element: lazy(<CreateCustomer />) },
      { path: 'customers/import', element: lazy(<ImportCustomers />) },
      { path: 'customers/merge', element: lazy(<MergeCustomers />) },
      { path: 'customers/:id', element: lazy(<CustomerProfile />) },

      // Loyalty
      { path: 'loyalty', element: lazy(<Loyalty />) },
      { path: 'loyalty/members', element: lazy(<LoyaltyMembers />) },
      { path: 'loyalty/rewards', element: lazy(<LoyaltyRewards />) },
      { path: 'loyalty/tiers', element: lazy(<LoyaltyTiers />) },

      // Campaigns
      { path: 'campaigns', element: lazy(<Campaigns />) },
      { path: 'campaigns/email', element: lazy(<EmailCampaign />) },
      { path: 'campaigns/sms', element: lazy(<SMSCampaign />) },
      { path: 'campaigns/:id', element: lazy(<CampaignDetail />) },

      // Segments
      { path: 'segments', element: lazy(<Segments />) },
      { path: 'segments/new', element: lazy(<SegmentBuilder />) },
      { path: 'segments/:id', element: lazy(<SegmentDetail />) },

      // Automations
      { path: 'automations', element: lazy(<Automations />) },
      { path: 'automations/new', element: lazy(<AutomationBuilder />) },

      // Templates, feedback, settings
      { path: 'templates', element: lazy(<Templates />) },
      { path: 'feedback', element: lazy(<Feedback />) },
      { path: 'settings', element: lazy(<SettingsScreen />) },

      // Sales Pipeline (add-on gated in sidebar; routes always present for deep links)
      { path: 'sales', element: lazy(<Pipeline />) },
      { path: 'sales/leads', element: lazy(<Leads />) },
      { path: 'sales/leads/:id', element: lazy(<LeadDetail />) },
      { path: 'sales/contracts', element: lazy(<SalesContracts />) },
    ],
  },
];

export default AppRoutes;
