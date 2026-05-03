import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const STATUS_ENUM = ['draft', 'active', 'paused', 'completed'];

export const listCampaigns: Tool = {
  name: 'crm_list_campaigns',
  description: 'List marketing campaigns. Filter by status.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: STATUS_ENUM, description: 'draft | active | paused | completed' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/crm/campaigns${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getCampaign: Tool = {
  name: 'crm_get_campaign',
  description: 'Get a single campaign by ID, including stats.',
  inputSchema: {
    type: 'object',
    properties: {
      campaign_id: { type: 'string', description: 'Campaign ID' },
    },
    required: ['campaign_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/crm/campaigns/${input.campaign_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createCampaign: Tool = {
  name: 'crm_create_campaign',
  description: 'Create a new marketing campaign.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Campaign name' },
      type: { type: 'string', description: 'email | sms | push' },
      subject: { type: 'string', description: 'Email subject line' },
      body: { type: 'string', description: 'Campaign message body' },
      segment_id: { type: 'string', description: 'Target audience segment ID' },
      scheduled_at: { type: 'string', description: 'ISO 8601 send time (omit to save as draft)' },
    },
    required: ['name', 'type'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/crm/campaigns', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateCampaign: Tool = {
  name: 'crm_update_campaign',
  description: 'Update a campaign (only possible while in draft status).',
  inputSchema: {
    type: 'object',
    properties: {
      campaign_id: { type: 'string', description: 'Campaign ID' },
      name: { type: 'string' },
      subject: { type: 'string' },
      body: { type: 'string' },
      scheduled_at: { type: 'string' },
    },
    required: ['campaign_id'],
  },
  handler: async (input, api) => {
    try {
      const { campaign_id, ...body } = input;
      const res = await api.put(`/crm/campaigns/${campaign_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
