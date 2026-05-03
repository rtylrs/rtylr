import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const listLoyaltyTiers: Tool = {
  name: 'crm_list_loyalty_tiers',
  description: 'List loyalty program tiers (Bronze, Silver, Gold, etc.).',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/crm/loyalty/tiers');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createLoyaltyTier: Tool = {
  name: 'crm_create_loyalty_tier',
  description: 'Create a new loyalty tier.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Tier name (e.g. Gold)' },
      points_required: { type: 'string', description: 'Minimum points to reach this tier' },
      discount_percent: { type: 'string', description: 'Discount percentage for members of this tier' },
      description: { type: 'string' },
    },
    required: ['name', 'points_required'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/crm/loyalty/tiers', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listLoyaltyMembers: Tool = {
  name: 'crm_list_loyalty_members',
  description: 'List customers enrolled in the loyalty program.',
  inputSchema: {
    type: 'object',
    properties: {
      tier_id: { type: 'string', description: 'Filter by tier ID' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const params = new URLSearchParams();
      if (input.tier_id) params.set('tier_id', String(input.tier_id));
      if (input.limit) params.set('limit', String(input.limit));
      if (input.offset) params.set('offset', String(input.offset));
      const q = params.toString() ? `?${params.toString()}` : '';
      const res = await api.get(`/crm/loyalty/members${q}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const assignLoyaltyTier: Tool = {
  name: 'crm_assign_loyalty_tier',
  description: 'Assign or change a customer\'s loyalty tier.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
      tier_id: { type: 'string', description: 'Loyalty tier ID' },
    },
    required: ['customer_id', 'tier_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`/crm/loyalty/members/${input.customer_id}/tier`, { tier_id: input.tier_id });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listLoyaltyRewards: Tool = {
  name: 'crm_list_loyalty_rewards',
  description: 'List redeemable loyalty rewards.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/crm/loyalty/rewards');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
