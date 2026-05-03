import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listCustomers: Tool = {
  name: 'crm_list_customers',
  description: 'List CRM customers. Supports search by name or email.',
  inputSchema: {
    type: 'object',
    properties: {
      search: { type: 'string', description: 'Search by name or email' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/crm/customers${qs(input, ['search', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const getCustomer: Tool = {
  name: 'crm_get_customer',
  description: 'Get a CRM customer by ID, including their loyalty status and purchase history.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
    },
    required: ['customer_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/crm/customers/${input.customer_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createCustomer: Tool = {
  name: 'crm_create_customer',
  description: 'Add a new customer to the CRM.',
  inputSchema: {
    type: 'object',
    properties: {
      name: { type: 'string', description: 'Full name' },
      email: { type: 'string', description: 'Email address' },
      phone: { type: 'string', description: 'Phone number' },
      notes: { type: 'string' },
    },
    required: ['name'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/crm/customers', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const updateCustomer: Tool = {
  name: 'crm_update_customer',
  description: 'Update a customer record.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
      name: { type: 'string' },
      email: { type: 'string' },
      phone: { type: 'string' },
      notes: { type: 'string' },
    },
    required: ['customer_id'],
  },
  handler: async (input, api) => {
    try {
      const { customer_id, ...body } = input;
      const res = await api.put(`/crm/customers/${customer_id}`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteCustomer: Tool = {
  name: 'crm_delete_customer',
  description: 'Delete a customer from the CRM.',
  inputSchema: {
    type: 'object',
    properties: {
      customer_id: { type: 'string', description: 'Customer ID' },
    },
    required: ['customer_id'],
  },
  handler: async (input, api) => {
    try {
      await api.delete(`/crm/customers/${input.customer_id}`);
      return ok({ deleted: true, customer_id: input.customer_id });
    } catch (e) {
      return apiErr(e);
    }
  },
};
