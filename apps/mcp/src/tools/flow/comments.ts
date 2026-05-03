import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const listComments: Tool = {
  name: 'flow_list_comments',
  description: 'List all comments on a task.',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
    },
    required: ['task_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/tasks/${input.task_id}/comments`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createComment: Tool = {
  name: 'flow_create_comment',
  description: 'Add a comment to a task.',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
      body: { type: 'string', description: 'Comment text (markdown supported)' },
    },
    required: ['task_id', 'body'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`/flow/tasks/${input.task_id}/comments`, { body: input.body });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
