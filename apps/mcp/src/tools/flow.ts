import { z } from 'zod';
import { createClient, apiError } from '../client.js';

export const flowTools = [
  {
    name: 'flow_list_workspaces',
    description: 'List all workspaces the authenticated user has access to in rtylr Flow.',
    inputSchema: {
      type: 'object' as const,
      properties: {},
      required: [],
    },
    handler: async (_input: Record<string, unknown>) => {
      try {
        const res = await createClient().get('/flow/workspaces');
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_list_projects',
    description: 'List projects in a workspace.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        workspace_id: { type: 'string', description: 'Workspace ID' },
      },
      required: ['workspace_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { workspace_id } = z.object({ workspace_id: z.string() }).parse(input);
        const res = await createClient().get(`/flow/projects?workspace_id=${workspace_id}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_list_tasks',
    description: 'List tasks in a project or workspace.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        project_id: { type: 'string', description: 'Project ID (optional if workspace_id given)' },
        workspace_id: { type: 'string', description: 'Workspace ID' },
        status: { type: 'string', description: 'Filter by status: todo, in_progress, in_review, done' },
        limit: { type: 'number', description: 'Max results (default 50)' },
        offset: { type: 'number', description: 'Pagination offset' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.project_id) params.set('project_id', String(input.project_id));
        if (input.workspace_id) params.set('workspace_id', String(input.workspace_id));
        if (input.status) params.set('status', String(input.status));
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/flow/tasks?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_get_task',
    description: 'Get a single task by ID.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        task_id: { type: 'string', description: 'Task ID' },
      },
      required: ['task_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { task_id } = z.object({ task_id: z.string() }).parse(input);
        const res = await createClient().get(`/flow/tasks/${task_id}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_create_task',
    description: 'Create a new task in rtylr Flow.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        title: { type: 'string', description: 'Task title (required)' },
        workspace_id: { type: 'string', description: 'Workspace ID (required)' },
        project_id: { type: 'string', description: 'Project ID (optional)' },
        status: { type: 'string', description: 'todo | in_progress | in_review | done' },
        priority: { type: 'string', description: 'none | low | medium | high | urgent' },
        assignee: { type: 'string', description: 'Assignee name' },
        due_date: { type: 'string', description: 'ISO 8601 date string' },
        description: { type: 'string', description: 'Task description (markdown)' },
      },
      required: ['title', 'workspace_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const schema = z.object({
          title: z.string(),
          workspace_id: z.string(),
          project_id: z.string().optional(),
          status: z.string().optional(),
          priority: z.string().optional(),
          assignee: z.string().optional(),
          due_date: z.string().optional(),
          description: z.string().optional(),
        });
        const data = schema.parse(input);
        const res = await createClient().post('/flow/tasks', data);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_update_task',
    description: 'Update an existing task.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        task_id: { type: 'string', description: 'Task ID to update (required)' },
        title: { type: 'string' },
        status: { type: 'string', description: 'todo | in_progress | in_review | done' },
        priority: { type: 'string', description: 'none | low | medium | high | urgent' },
        assignee: { type: 'string' },
        due_date: { type: 'string' },
        description: { type: 'string' },
      },
      required: ['task_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { task_id, ...rest } = z.object({
          task_id: z.string(),
          title: z.string().optional(),
          status: z.string().optional(),
          priority: z.string().optional(),
          assignee: z.string().optional(),
          due_date: z.string().optional(),
          description: z.string().optional(),
        }).parse(input);
        const res = await createClient().put(`/flow/tasks/${task_id}`, rest);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_delete_task',
    description: 'Delete a task permanently.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        task_id: { type: 'string', description: 'Task ID to delete' },
      },
      required: ['task_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { task_id } = z.object({ task_id: z.string() }).parse(input);
        await createClient().delete(`/flow/tasks/${task_id}`);
        return { content: [{ type: 'text', text: `Task ${task_id} deleted.` }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_create_project',
    description: 'Create a new project in a workspace.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        name: { type: 'string', description: 'Project name (required)' },
        workspace_id: { type: 'string', description: 'Workspace ID (required)' },
        description: { type: 'string' },
        color: { type: 'string', description: 'Hex color e.g. #64C5B9' },
      },
      required: ['name', 'workspace_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const data = z.object({
          name: z.string(),
          workspace_id: z.string(),
          description: z.string().optional(),
          color: z.string().optional(),
        }).parse(input);
        const res = await createClient().post('/flow/projects', data);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'flow_my_tasks',
    description: "List tasks assigned to the authenticated user.",
    inputSchema: {
      type: 'object' as const,
      properties: {
        status: { type: 'string', description: 'Filter by status' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.status) params.set('status', String(input.status));
        const res = await createClient().get(`/flow/tasks/mine?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
];
