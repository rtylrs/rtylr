import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const STATUS_ENUM = ['draft', 'pending', 'paid', 'overdue', 'cancelled'];

export const listBills: Tool = {
  name: 'finance_list_bills',
  description: 'List accounts-payable bills.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: STATUS_ENUM, description: 'draft | pending | paid | overdue | cancelled' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/bills${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getBill: Tool = {
  name: 'finance_get_bill',
  description: 'Get a single bill by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      bill_id: { type: 'string', description: 'Bill ID' },
    },
    required: ['bill_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/bills/${input.bill_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createBill: Tool = {
  name: 'finance_create_bill',
  description: 'Create a new bill (accounts payable).',
  inputSchema: {
    type: 'object',
    properties: {
      supplier_id: { type: 'string', description: 'Supplier ID' },
      amount: { type: 'string', description: 'Amount in smallest currency unit' },
      currency: { type: 'string', description: 'ISO currency code e.g. USD' },
      due_date: { type: 'string', description: 'ISO 8601 date' },
      notes: { type: 'string' },
    },
    required: ['supplier_id', 'amount'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/finance/bills', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateBill: Tool = {
  name: 'finance_update_bill',
  description: 'Update a bill.',
  inputSchema: {
    type: 'object',
    properties: {
      bill_id: { type: 'string', description: 'Bill ID' },
      status: { type: 'string', enum: STATUS_ENUM },
      due_date: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['bill_id'],
  },
  handler: async (input, api) => {
    try {
      const { bill_id, ...body } = input;
      const res = await api.put(`/finance/bills/${bill_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteBill: Tool = {
  name: 'finance_delete_bill',
  description: 'Delete a bill.',
  inputSchema: {
    type: 'object',
    properties: {
      bill_id: { type: 'string', description: 'Bill ID' },
    },
    required: ['bill_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`/finance/bills/${input.bill_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
