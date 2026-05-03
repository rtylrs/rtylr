import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const inventoryReport: Tool = {
  name: 'erp_inventory_report',
  description: 'Get current inventory levels and stock report.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/erp/inventory/report');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listPriceLists: Tool = {
  name: 'erp_list_price_lists',
  description: 'List ERP price lists.',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/erp/price-lists');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createPriceList: Tool = {
  name: 'erp_create_price_list',
  description: 'Create a new price list.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Price list name' },
      description: { type: 'string' },
      currency: { type: 'string', description: 'ISO currency code' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/erp/price-lists', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listStockTransfers: Tool = {
  name: 'erp_list_stock_transfers',
  description: 'List stock transfers between locations.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', description: 'pending | in_transit | completed | cancelled' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/erp/transfers${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listStockTakes: Tool = {
  name: 'erp_list_stock_takes',
  description: 'List stock take / physical inventory count sessions.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', description: 'open | completed' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/erp/stock-takes${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
