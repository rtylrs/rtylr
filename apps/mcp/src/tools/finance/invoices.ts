import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const STATUS_ENUM = ['draft', 'sent', 'paid', 'overdue', 'cancelled'];

export const listInvoices: Tool = {
  name: 'finance_list_invoices',
  description: 'List finance invoices. Filter by status.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: STATUS_ENUM, description: 'draft | sent | paid | overdue | cancelled' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/invoices${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getInvoice: Tool = {
  name: 'finance_get_invoice',
  description: 'Get a single invoice by ID.',
  inputSchema: {
    type: 'object',
    properties: {
      invoice_id: { type: 'string', description: 'Invoice ID' },
    },
    required: ['invoice_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/invoices/${input.invoice_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createInvoice: Tool = {
  name: 'finance_create_invoice',
  description: 'Create a new invoice.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
      amount: { type: 'string', description: 'Total amount in smallest currency unit' },
      currency: { type: 'string', description: 'ISO currency code e.g. USD' },
      due_date: { type: 'string', description: 'ISO 8601 date' },
      notes: { type: 'string' },
    },
    required: ['customer_id', 'amount'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/finance/invoices', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateInvoice: Tool = {
  name: 'finance_update_invoice',
  description: 'Update an invoice.',
  inputSchema: {
    type: 'object',
    properties: {
      invoice_id: { type: 'string', description: 'Invoice ID' },
      status: { type: 'string', enum: STATUS_ENUM },
      due_date: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['invoice_id'],
  },
  handler: async (input, api) => {
    try {
      const { invoice_id, ...body } = input;
      const res = await api.put(`/finance/invoices/${invoice_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteInvoice: Tool = {
  name: 'finance_delete_invoice',
  description: 'Delete an invoice (only possible while in draft status).',
  inputSchema: {
    type: 'object',
    properties: {
      invoice_id: { type: 'string', description: 'Invoice ID' },
    },
    required: ['invoice_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`/finance/invoices/${input.invoice_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
