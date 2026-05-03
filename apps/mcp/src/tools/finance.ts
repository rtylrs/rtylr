import { z } from 'zod';
import { createClient, apiError } from '../client.js';

export const financeTools = [
  {
    name: 'finance_list_invoices',
    description: 'List finance invoices.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        status: { type: 'string', description: 'draft | sent | paid | overdue | cancelled' },
        limit: { type: 'number' },
        offset: { type: 'number' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.status) params.set('status', String(input.status));
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/finance/invoices?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'finance_list_expenses',
    description: 'List finance expenses.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        limit: { type: 'number' },
        offset: { type: 'number' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/finance/expenses?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'finance_profit_loss',
    description: 'Get profit & loss report for a date range.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        from: { type: 'string', description: 'Start date YYYY-MM-DD (required)' },
        to: { type: 'string', description: 'End date YYYY-MM-DD (required)' },
      },
      required: ['from', 'to'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { from, to } = z.object({ from: z.string(), to: z.string() }).parse(input);
        const res = await createClient().get(`/finance/reports/profit-loss?from=${from}&to=${to}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'finance_create_invoice',
    description: 'Create a new invoice.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        customer_id: { type: 'string', description: 'Customer ID (required)' },
        amount: { type: 'number', description: 'Total amount in smallest currency unit (required)' },
        currency: { type: 'string', description: 'ISO currency code e.g. USD' },
        due_date: { type: 'string', description: 'ISO 8601 date' },
        notes: { type: 'string' },
      },
      required: ['customer_id', 'amount'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const data = z.object({
          customer_id: z.string(),
          amount: z.number(),
          currency: z.string().optional(),
          due_date: z.string().optional(),
          notes: z.string().optional(),
        }).parse(input);
        const res = await createClient().post('/finance/invoices', data);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
];
