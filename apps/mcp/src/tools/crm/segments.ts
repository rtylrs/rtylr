import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const listSegments: Tool = {
  name: 'crm_list_segments',
  description: 'List customer segments (audience groups used for targeting campaigns).',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/crm/segments');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createSegment: Tool = {
  name: 'crm_create_segment',
  description: 'Create a new customer segment with filter rules.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Segment name' },
      description: { type: 'string' },
      rules: { type: 'string', description: 'JSON-encoded filter rules array' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const body: Record<string, unknown> = { name: input.name, description: input.description };
      if (input.rules) {
        try {
          body.rules = JSON.parse(input.rules as string);
        } catch {
          body.rules = [];
        }
      }
      const res = await api.post('/crm/segments', body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
