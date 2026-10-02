import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const listWorkspaces: Tool = {
  name: 'flow_list_workspaces',
  description: 'List all Flow workspaces the authenticated user has access to.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/flow/workspaces');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createWorkspace: Tool = {
  name: 'flow_create_workspace',
  description: 'Create a new Flow workspace.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Workspace name' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/flow/workspaces', { name: input.name });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
