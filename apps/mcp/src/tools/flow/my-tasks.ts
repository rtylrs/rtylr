import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const myTasks: Tool = {
  name: 'flow_my_tasks',
  description: 'List tasks assigned to the authenticated user, grouped or filtered by status.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', description: 'Filter by status: todo | in_progress | in_review | done' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/tasks/mine${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
