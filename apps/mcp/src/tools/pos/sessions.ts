import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listSessions: Tool = {
  name: 'pos_list_sessions',
  description: 'List POS sessions. Supports filtering by location and status.',
  inputSchema: {
    type: 'object',
    properties: {
      location_id: { type: 'string', description: 'Filter by location ID' },
      status: { type: 'string', description: 'Filter by status (open or closed)', enum: ['open', 'closed'] },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/sessions${qs(input, ['location_id', 'status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getSession: Tool = {
  name: 'pos_get_session',
  description: 'Get a POS session by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      session_id: { type: 'string', description: 'Session ID' },
    },
    required: ['session_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/sessions/${input.session_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const openSession: Tool = {
  name: 'pos_open_session',
  description: 'Open a new POS session.',
  inputSchema: {
    type: 'object',
    properties: {
      location_id: { type: 'string', description: 'Location ID' },
      opening_float: { type: 'string', description: 'Opening float amount in cents' },
    },
    required: ['location_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/pos/sessions', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const closeSession: Tool = {
  name: 'pos_close_session',
  description: 'Close a POS session.',
  inputSchema: {
    type: 'object',
    properties: {
      session_id: { type: 'string', description: 'Session ID' },
      closing_float: { type: 'string', description: 'Closing float amount in cents' },
    },
    required: ['session_id'],
  },
  handler: async (input, api) => {
    try {
      const { session_id, ...body } = input;
      const res = await api.post(`/pos/sessions/${session_id}/close`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
