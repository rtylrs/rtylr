import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listMenuItems: Tool = {
  name: 'pos_list_menu_items',
  description: 'List POS menu items. Supports filtering by category and search.',
  inputSchema: {
    type: 'object',
    properties: {
      category_id: { type: 'string', description: 'Filter by category ID' },
      search: { type: 'string', description: 'Search by name' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/menu-items${qs(input, ['category_id', 'search', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getMenuItem: Tool = {
  name: 'pos_get_menu_item',
  description: 'Get a POS menu item by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      item_id: { type: 'string', description: 'Menu item ID' },
    },
    required: ['item_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/menu-items/${input.item_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createMenuItem: Tool = {
  name: 'pos_create_menu_item',
  description: 'Create a new POS menu item.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Item name' },
      price: { type: 'string', description: 'Price in smallest currency unit (cents)' },
      category_id: { type: 'string', description: 'Category ID' },
      description: { type: 'string', description: 'Item description' },
      available: { type: 'string', description: 'Whether the item is available ("true" or "false")' },
    },
    required: ['name', 'price'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/pos/menu-items', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateMenuItem: Tool = {
  name: 'pos_update_menu_item',
  description: 'Update a POS menu item.',
  inputSchema: {
    type: 'object',
    properties: {
      item_id: { type: 'string', description: 'Menu item ID' },
      name: { type: 'string', description: 'Item name' },
      price: { type: 'string', description: 'Price in smallest currency unit (cents)' },
      available: { type: 'string', description: 'Whether the item is available ("true" or "false")' },
      description: { type: 'string', description: 'Item description' },
    },
    required: ['item_id'],
  },
  handler: async (input, api) => {
    try {
      const { item_id, ...body } = input;
      const res = await api.put(`/pos/menu-items/${item_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const toggleMenuItemAvailability: Tool = {
  name: 'pos_toggle_menu_item_availability',
  description: 'Toggle the availability of a POS menu item.',
  inputSchema: {
    type: 'object',
    properties: {
      item_id: { type: 'string', description: 'Menu item ID' },
      available: { type: 'string', description: 'Availability: "true" or "false"' },
    },
    required: ['item_id', 'available'],
  },
  handler: async (input, api) => {
    try {
      const { item_id, available } = input;
      const res = await api.put(`/pos/menu-items/${item_id}/availability`, { available });
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
