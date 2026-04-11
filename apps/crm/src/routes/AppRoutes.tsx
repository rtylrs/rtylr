import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout';
import Dashboard from '../features/dashboard/Dashboard';
import Customers from '../features/customers/Customers';
import CustomerProfile from '../features/customers/CustomerProfile';
import CreateCustomer from '../features/customers/CreateCustomer';
import ImportCustomers from '../features/customers/ImportCustomers';
import MergeCustomers from '../features/customers/MergeCustomers';
import Loyalty from '../features/loyalty/Loyalty';
import LoyaltyMembers from '../features/loyalty/LoyaltyMembers';
import LoyaltyRewards from '../features/loyalty/LoyaltyRewards';
import LoyaltyTiers from '../features/loyalty/LoyaltyTiers';
import Campaigns from '../features/campaigns/Campaigns';
import EmailCampaign from '../features/campaigns/EmailCampaign';
import SMSCampaign from '../features/campaigns/SMSCampaign';
import CampaignDetail from '../features/campaigns/CampaignDetail';
import Segments from '../features/segments/Segments';
import SegmentBuilder from '../features/segments/SegmentBuilder';
import SegmentDetail from '../features/segments/SegmentDetail';
import Automations from '../features/automations/Automations';
import AutomationBuilder from '../features/automations/AutomationBuilder';
import Templates from '../features/templates/Templates';
import Feedback from '../features/feedback/Feedback';
import SettingsScreen from '../features/settings/Settings';

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

const AppRoutes: AppRoute[] = [
  {
    path: '/',
    element: <DefaultLayout />,
    children: [
      { index: true, element: <Dashboard /> },

      // Customers
      { path: 'customers', element: <Customers /> },
      { path: 'customers/new', element: <CreateCustomer /> },
      { path: 'customers/import', element: <ImportCustomers /> },
      { path: 'customers/merge', element: <MergeCustomers /> },
      { path: 'customers/:id', element: <CustomerProfile /> },

      // Loyalty
      { path: 'loyalty', element: <Loyalty /> },
      { path: 'loyalty/members', element: <LoyaltyMembers /> },
      { path: 'loyalty/rewards', element: <LoyaltyRewards /> },
      { path: 'loyalty/tiers', element: <LoyaltyTiers /> },

      // Campaigns
      { path: 'campaigns', element: <Campaigns /> },
      { path: 'campaigns/email', element: <EmailCampaign /> },
      { path: 'campaigns/sms', element: <SMSCampaign /> },
      { path: 'campaigns/:id', element: <CampaignDetail /> },

      // Segments
      { path: 'segments', element: <Segments /> },
      { path: 'segments/new', element: <SegmentBuilder /> },
      { path: 'segments/:id', element: <SegmentDetail /> },

      // Automations
      { path: 'automations', element: <Automations /> },
      { path: 'automations/new', element: <AutomationBuilder /> },

      // Templates, feedback, settings
      { path: 'templates', element: <Templates /> },
      { path: 'feedback', element: <Feedback /> },
      { path: 'settings', element: <SettingsScreen /> },
    ],
  },
];

export default AppRoutes;
