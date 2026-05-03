import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const listExpenses: Tool = {
  name: 'finance_list_expenses',
  description: 'List finance expenses.',
  inputSchema: {
    type: 'object',
    properties: {
      category: { type: 'string', description: 'Filter by category' },
      from: { type: 'string', description: 'Start date YYYY-MM-DD' },
      to: { type: 'string', description: 'End date YYYY-MM-DD' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/expenses${qs(input, ['category', 'from', 'to', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const createExpense: Tool = {
  name: 'finance_create_expense',
  description: 'Record a new expense.',
  inputSchema: {
    type: 'object',
    properties: {
      amount: { type: 'string', description: 'Amount in smallest currency unit' },
      currency: { type: 'string', description: 'ISO currency code e.g. USD' },
      category: { type: 'string', description: 'Expense category' },
      description: { type: 'string' },
      date: { type: 'string', description: 'ISO 8601 date' },
    },
    required: ['amount', 'category'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post('/finance/expenses', input);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const deleteExpense: Tool = {
  name: 'finance_delete_expense',
  description: 'Delete an expense.',
  inputSchema: {
    type: 'object',
    properties: {
      expense_id: { type: 'string', description: 'Expense ID' },
    },
    required: ['expense_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.delete(`/finance/expenses/${input.expense_id}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
