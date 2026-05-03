import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

const STATUS_ENUM = ['draft', 'sent', 'accepted', 'rejected', 'expired', 'converted'];

export const listQuotes: Tool = {
  name: 'finance_list_quotes',
  description: 'List finance quotes / estimates.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', enum: STATUS_ENUM, description: 'draft | sent | accepted | rejected | expired | converted' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/quotes${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createQuote: Tool = {
  name: 'finance_create_quote',
  description: 'Create a new quote.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
      amount: { type: 'string', description: 'Total amount in smallest currency unit' },
      currency: { type: 'string', description: 'ISO currency code e.g. USD' },
      valid_until: { type: 'string', description: 'ISO 8601 expiry date' },
      notes: { type: 'string' },
    },
    required: ['customer_id', 'amount'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/finance/quotes', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateQuote: Tool = {
  name: 'finance_update_quote',
  description: 'Update a quote.',
  inputSchema: {
    type: 'object',
    properties: {
      quote_id: { type: 'string', description: 'Quote ID' },
      status: { type: 'string', enum: STATUS_ENUM },
      valid_until: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['quote_id'],
  },
  handler: async (input, api) => {
    try {
      const { quote_id, ...body } = input;
      const res = await api.put(`/finance/quotes/${quote_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteQuote: Tool = {
  name: 'finance_delete_quote',
  description: 'Delete a quote.',
  inputSchema: {
    type: 'object',
    properties: {
      quote_id: { type: 'string', description: 'Quote ID' },
    },
    required: ['quote_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`/finance/quotes/${input.quote_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const convertQuote: Tool = {
  name: 'finance_convert_quote',
  description: 'Convert an accepted quote into an invoice.',
  inputSchema: {
    type: 'object',
    properties: {
      quote_id: { type: 'string', description: 'Quote ID' },
    },
    required: ['quote_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`/finance/quotes/${input.quote_id}/convert`, {});
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
