import type { Tool } from '../types.js';
import { ok, apiErr, err, qs } from '../types.js';

const PO_BASE = '/api/erp/purchase-order/v1/purchase-orders';
const STATUS_ENUM = ['draft', 'sent', 'approved', 'partial', 'received', 'cancelled'];

export const listPurchaseOrders: Tool = {
  name: 'erp_list_purchase_orders',
  description: 'List ERP purchase orders.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: STATUS_ENUM, description: 'draft | sent | approved | partial | received | cancelled' },
      supplier_id: { type: 'string' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`${PO_BASE}${qs(input, ['status', 'supplier_id', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getPurchaseOrder: Tool = {
  name: 'erp_get_purchase_order',
  description: 'Get a single purchase order by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      po_id: { type: 'string', description: 'Purchase order ID' },
    },
    required: ['po_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`${PO_BASE}/${encodeURIComponent(String(input.po_id))}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createPurchaseOrder: Tool = {
  name: 'erp_create_purchase_order',
  description: 'Create a new purchase order.',
  inputSchema: {
    type: 'object',
    properties: {
      supplier_id: { type: 'string', description: 'Supplier ID' },
      expected_date: { type: 'string', description: 'ISO 8601 expected delivery date' },
      notes: { type: 'string' },
      items: { type: 'string', description: 'JSON array of {product_id, quantity, unit_price}' },
    },
    required: ['supplier_id'],
  },
  handler: async (input, api) => {
    try {
      const body: Record<string, unknown> = { ...input };
      if (typeof input.items === 'string') {
        try {
          body.items = JSON.parse(input.items as string);
        } catch {
          return err('Invalid items: must be a JSON array of {product_id, quantity, unit_price}');
        }
      }
      const res = await api.post(PO_BASE, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const approvePurchaseOrder: Tool = {
  name: 'erp_approve_purchase_order',
  description: 'Approve a purchase order.',
  inputSchema: {
    type: 'object',
    properties: {
      po_id: { type: 'string', description: 'Purchase order ID' },
    },
    required: ['po_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`${PO_BASE}/${encodeURIComponent(String(input.po_id))}/approve`, {});
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
