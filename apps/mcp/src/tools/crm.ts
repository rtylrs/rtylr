import { z } from 'zod';
import { createClient, apiError } from '../client.js';

export const crmTools = [
  {
    name: 'crm_list_customers',
    description: 'List CRM customers.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        search: { type: 'string', description: 'Search by name or email' },
        limit: { type: 'number', description: 'Max results (default 50)' },
        offset: { type: 'number' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.search) params.set('search', String(input.search));
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/crm/customers?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'crm_get_customer',
    description: 'Get a CRM customer by ID.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        customer_id: { type: 'string' },
      },
      required: ['customer_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { customer_id } = z.object({ customer_id: z.string() }).parse(input);
        const res = await createClient().get(`/crm/customers/${customer_id}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'crm_list_campaigns',
    description: 'List CRM campaigns.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        status: { type: 'string', description: 'draft | active | paused | completed' },
        limit: { type: 'number' },
        offset: { type: 'number' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.status) params.set('status', String(input.status));
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/crm/campaigns?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'crm_list_loyalty_tiers',
    description: 'List loyalty program tiers.',
    inputSchema: {
      type: 'object' as const,
      properties: {},
      required: [],
    },
    handler: async (_input: Record<string, unknown>) => {
      try {
        const res = await createClient().get('/crm/loyalty/tiers');
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
];
