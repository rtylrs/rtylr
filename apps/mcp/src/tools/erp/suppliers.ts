import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listSuppliers: Tool = {
  name: 'erp_list_suppliers',
  description: 'List ERP suppliers.',
  inputSchema: {
    type: 'object',
    properties: {
      search: { type: 'string', description: 'Search by name' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/erp/suppliers${qs(input, ['search', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createSupplier: Tool = {
  name: 'erp_create_supplier',
  description: 'Create a new supplier.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Supplier name' },
      email: { type: 'string' },
      phone: { type: 'string' },
      address: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/erp/suppliers', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateSupplier: Tool = {
  name: 'erp_update_supplier',
  description: 'Update a supplier.',
  inputSchema: {
    type: 'object',
    properties: {
      supplier_id: { type: 'string', description: 'Supplier ID' },
      name: { type: 'string' },
      email: { type: 'string' },
      phone: { type: 'string' },
      address: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['supplier_id'],
  },
  handler: async (input, api) => {
    try {
      const { supplier_id, ...body } = input;
      const res = await api.put(`/erp/suppliers/${supplier_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteSupplier: Tool = {
  name: 'erp_delete_supplier',
  description: 'Delete a supplier.',
  inputSchema: {
    type: 'object',
    properties: {
      supplier_id: { type: 'string', description: 'Supplier ID' },
    },
    required: ['supplier_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`/erp/suppliers/${input.supplier_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
