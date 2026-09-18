export { listCustomers, getCustomer, createCustomer, updateCustomer, deleteCustomer } from './customers.js';
export { listCampaigns, getCampaign, createCampaign, updateCampaign } from './campaigns.js';
export { listSegments, createSegment } from './segments.js';
export { listLoyaltyTiers, createLoyaltyTier, listLoyaltyMembers, assignLoyaltyTier, listLoyaltyRewards } from './loyalty.js';
export { listAutomations, toggleAutomation } from './automations.js';
export { listFeedback } from './feedback.js';

import { listCustomers, getCustomer, createCustomer, updateCustomer, deleteCustomer } from './customers.js';
import { listCampaigns, getCampaign, createCampaign, updateCampaign } from './campaigns.js';
import { listSegments, createSegment } from './segments.js';
import { listLoyaltyTiers, createLoyaltyTier, listLoyaltyMembers, assignLoyaltyTier, listLoyaltyRewards } from './loyalty.js';
import { listAutomations, toggleAutomation } from './automations.js';
import { listFeedback } from './feedback.js';

export const crmTools = [
  listCustomers, getCustomer, createCustomer, updateCustomer, deleteCustomer,
  listCampaigns, getCampaign, createCampaign, updateCampaign,
  listSegments, createSegment,
  listLoyaltyTiers, createLoyaltyTier, listLoyaltyMembers, assignLoyaltyTier, listLoyaltyRewards,
  listAutomations, toggleAutomation,
  listFeedback,
];
