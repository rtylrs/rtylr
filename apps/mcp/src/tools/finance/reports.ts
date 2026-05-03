import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const profitLoss: Tool = {
  name: 'finance_profit_loss',
  description: 'Get profit & loss report for a date range.',
  inputSchema: {
    type: 'object',
    properties: {
      from: { type: 'string', description: 'Start date YYYY-MM-DD' },
      to: { type: 'string', description: 'End date YYYY-MM-DD' },
    },
    required: ['from', 'to'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/reports/profit-loss${qs(input, ['from', 'to'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const balanceSheet: Tool = {
  name: 'finance_balance_sheet',
  description: 'Get balance sheet as of a given date.',
  inputSchema: {
    type: 'object',
    properties: {
      as_of: { type: 'string', description: 'Date YYYY-MM-DD (defaults to today)' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/reports/balance-sheet${qs(input, ['as_of'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const cashFlow: Tool = {
  name: 'finance_cash_flow',
  description: 'Get cash flow statement for a date range.',
  inputSchema: {
    type: 'object',
    properties: {
      from: { type: 'string', description: 'Start date YYYY-MM-DD' },
      to: { type: 'string', description: 'End date YYYY-MM-DD' },
    },
    required: ['from', 'to'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/reports/cash-flow${qs(input, ['from', 'to'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const taxConfig: Tool = {
  name: 'finance_tax_config',
  description: 'Get tax configuration (rates and rules).',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/finance/tax/config');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const chartOfAccounts: Tool = {
  name: 'finance_chart_of_accounts',
  description: 'Get the chart of accounts (hierarchical tree).',
  inputSchema: { type: 'object', properties: {} },
  handler: async (_input, api) => {
    try {
      const res = await api.get('/finance/chart-of-accounts');
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const listJournalEntries: Tool = {
  name: 'finance_list_journal_entries',
  description: 'List journal entries.',
  inputSchema: {
    type: 'object',
    properties: {
      status: { type: 'string', description: 'draft | posted' },
      limit: { type: 'string' },
      offset: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/finance/journal${qs(input, ['status', 'limit', 'offset'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const postJournalEntry: Tool = {
  name: 'finance_post_journal_entry',
  description: 'Post a draft journal entry (moves it to posted status).',
  inputSchema: {
    type: 'object',
    properties: {
      entry_id: { type: 'string', description: 'Journal entry ID' },
    },
    required: ['entry_id'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.post(`/finance/journal/${input.entry_id}/post`, {});
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
