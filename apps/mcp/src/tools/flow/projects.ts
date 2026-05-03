import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listProjects: Tool = {
  name: 'flow_list_projects',
  description: 'List projects in a workspace.',
  inputSchema: {
    type: 'object',
    properties: {
      workspace_id: { type: 'string', description: 'Workspace ID' },
    },
    required: ['workspace_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/projects${qs(input, ['workspace_id'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getProject: Tool = {
  name: 'flow_get_project',
  description: 'Get a single project by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      project_id: { type: 'string', description: 'Project ID' },
    },
    required: ['project_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/projects/${input.project_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createProject: Tool = {
  name: 'flow_create_project',
  description: 'Create a new project in a workspace.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Project name' },
      workspace_id: { type: 'string', description: 'Workspace ID' },
      description: { type: 'string' },
      color: { type: 'string', description: 'Hex color, e.g. #64C5B9' },
    },
    required: ['name', 'workspace_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/flow/projects', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateProject: Tool = {
  name: 'flow_update_project',
  description: 'Update a project name, description, or color.',
  inputSchema: {
    type: 'object',
    properties: {
      project_id: { type: 'string', description: 'Project ID' },
      name: { type: 'string' },
      description: { type: 'string' },
      color: { type: 'string' },
    },
    required: ['project_id'],
  },
  handler: async (input, api) => {
    try {
      const { project_id, ...body } = input;
      const res = await api.put(`/flow/projects/${project_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteProject: Tool = {
  name: 'flow_delete_project',
  description: 'Delete a project and all its tasks permanently.',
  inputSchema: {
    type: 'object',
    properties: {
      project_id: { type: 'string', description: 'Project ID' },
    },
    required: ['project_id'],
  },
  handler: async (input, api) => {
    try {
      await api.delete(`/flow/projects/${input.project_id}`);
      return ok({ deleted: true, project_id: input.project_id });
    } catch (e) {
      return apiErr(e);
    }
  },
};
