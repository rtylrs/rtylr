import React from 'react';
import DefaultLayout from '../layouts/DefaultLayout/DefaultLayout';
import Dashboard from '../screens/Dashboard/Dashboard';
import Customers from '../screens/Customers/Customers';
import CustomerProfile from '../screens/Customers/CustomerProfile';
import CreateCustomer from '../screens/Customers/CreateCustomer';
import ImportCustomers from '../screens/Customers/ImportCustomers';
import MergeCustomers from '../screens/Customers/MergeCustomers';
import Loyalty from '../screens/Loyalty/Loyalty';
import LoyaltyMembers from '../screens/Loyalty/LoyaltyMembers';
import LoyaltyRewards from '../screens/Loyalty/LoyaltyRewards';
import LoyaltyTiers from '../screens/Loyalty/LoyaltyTiers';
import Campaigns from '../screens/Campaigns/Campaigns';
import EmailCampaign from '../screens/Campaigns/EmailCampaign';
import SMSCampaign from '../screens/Campaigns/SMSCampaign';
import CampaignDetail from '../screens/Campaigns/CampaignDetail';
import Segments from '../screens/Segments/Segments';
import SegmentBuilder from '../screens/Segments/SegmentBuilder';
import SegmentDetail from '../screens/Segments/SegmentDetail';
import Automations from '../screens/Automations/Automations';
import AutomationBuilder from '../screens/Automations/AutomationBuilder';
import Templates from '../screens/Templates/Templates';
import Feedback from '../screens/Feedback/Feedback';
import SettingsScreen from '../screens/Settings/Settings';

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
