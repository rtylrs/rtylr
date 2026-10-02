import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listFeedback: Tool = {
  name: 'crm_list_feedback',
  description: 'List customer feedback and reviews.',
  inputSchema: {
    type: 'object',
    properties: {
      rating: { type: 'string', description: 'Filter by rating (1-5)' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/crm/feedback${qs(input, ['rating', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
