import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listOrders: Tool = {
  name: 'pos_list_orders',
  description: 'List POS orders. Supports filtering by status.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', description: 'Filter by status (pending, preparing, ready, served, cancelled)' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/orders${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getOrder: Tool = {
  name: 'pos_get_order',
  description: 'Get a POS order by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      order_id: { type: 'string', description: 'Order ID' },
    },
    required: ['order_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/orders/${input.order_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createOrder: Tool = {
  name: 'pos_create_order',
  description: 'Create a new POS order.',
  inputSchema: {
    type: 'object',
    properties: {
      location_id: { type: 'string', description: 'Location ID' },
      items: { type: 'string', description: 'JSON array of items: [{menu_item_id, quantity, modifiers?}]' },
    },
    required: ['location_id'],
  },
  handler: async (input, api) => {
    try {
      const { location_id, items, ...rest } = input;
      const body: Record<string, unknown> = { location_id, ...rest };
      if (items) {
        try {
          body.items = JSON.parse(items as string);
        } catch {
          body.items = items;
        }
      }
      const res = await api.post('/pos/orders', body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateOrderStatus: Tool = {
  name: 'pos_update_order_status',
  description: 'Update the status of a POS order.',
  inputSchema: {
    type: 'object',
    properties: {
      order_id: { type: 'string', description: 'Order ID' },
      status: {
        type: 'string',
        description: 'New status',
        enum: ['pending', 'preparing', 'ready', 'served', 'cancelled'],
      },
    },
    required: ['order_id', 'status'],
  },
  handler: async (input, api) => {
    try {
      const { order_id, status } = input;
      const res = await api.put(`/pos/orders/${order_id}/status`, { status });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const voidOrder: Tool = {
  name: 'pos_void_order',
  description: 'Void a POS order.',
  inputSchema: {
    type: 'object',
    properties: {
      order_id: { type: 'string', description: 'Order ID' },
    },
    required: ['order_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`/pos/orders/${input.order_id}/void`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
