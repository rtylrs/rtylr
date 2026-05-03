import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const STATUS_ENUM = ['todo', 'in_progress', 'in_review', 'done'];
const PRIORITY_ENUM = ['none', 'low', 'medium', 'high', 'urgent'];

export const listTasks: Tool = {
  name: 'flow_list_tasks',
  description: 'List tasks in a project or workspace. Supports filtering by status and priority.',
  inputSchema: {
    type: 'object',
    properties: {
      project_id: { type: 'string', description: 'Filter by project ID' },
      workspace_id: { type: 'string', description: 'Filter by workspace ID' },
      status: { type: 'string', description: 'todo | in_progress | in_review | done', enum: STATUS_ENUM },
      priority: { type: 'string', description: 'none | low | medium | high | urgent', enum: PRIORITY_ENUM },
      assignee: { type: 'string', description: 'Filter by assignee name' },
      limit: { type: 'string', description: 'Max results (default 50)' },
      offset: { type: 'string', description: 'Pagination offset' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/tasks${qs(input, ['project_id', 'workspace_id', 'status', 'priority', 'assignee', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getTask: Tool = {
  name: 'flow_get_task',
  description: 'Get a single task by ID, including its description, labels, subtasks, and comments.',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
    },
    required: ['task_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/flow/tasks/${input.task_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createTask: Tool = {
  name: 'flow_create_task',
  description: 'Create a new task.',
  inputSchema: {
    type: 'object',
    properties: {
      title: { type: 'string', description: 'Task title' },
      workspace_id: { type: 'string', description: 'Workspace ID' },
      project_id: { type: 'string', description: 'Project ID (optional)' },
      status: { type: 'string', enum: STATUS_ENUM },
      priority: { type: 'string', enum: PRIORITY_ENUM },
      assignee: { type: 'string', description: 'Assignee name' },
      due_date: { type: 'string', description: 'ISO 8601 date' },
      description: { type: 'string', description: 'Task description (markdown supported)' },
    },
    required: ['title', 'workspace_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/flow/tasks', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateTask: Tool = {
  name: 'flow_update_task',
  description: 'Update an existing task — title, status, priority, assignee, due date, or description.',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
      title: { type: 'string' },
      status: { type: 'string', enum: STATUS_ENUM },
      priority: { type: 'string', enum: PRIORITY_ENUM },
      assignee: { type: 'string' },
      due_date: { type: 'string', description: 'ISO 8601 date' },
      description: { type: 'string' },
    },
    required: ['task_id'],
  },
  handler: async (input, api) => {
    try {
      const { task_id, ...body } = input;
      const res = await api.put(`/flow/tasks/${task_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteTask: Tool = {
  name: 'flow_delete_task',
  description: 'Permanently delete a task.',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
    },
    required: ['task_id'],
  },
  handler: async (input, api) => {
    try {
      await api.delete(`/flow/tasks/${input.task_id}`);
      return ok({ deleted: true, task_id: input.task_id });
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateTaskStatus: Tool = {
  name: 'flow_update_task_status',
  description: 'Quickly move a task to a new status (shortcut for flow_update_task).',
  inputSchema: {
    type: 'object',
    properties: {
      task_id: { type: 'string', description: 'Task ID' },
      status: { type: 'string', description: 'New status', enum: STATUS_ENUM },
    },
    required: ['task_id', 'status'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.put(`/flow/tasks/${input.task_id}`, { status: input.status });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
