import { z } from 'zod';
import { createClient, apiError } from '../client.js';

export const erpTools = [
  {
    name: 'erp_list_products',
    description: 'List ERP products/inventory items.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        search: { type: 'string' },
        limit: { type: 'number' },
        offset: { type: 'number' },
      },
      required: [],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const params = new URLSearchParams();
        if (input.search) params.set('search', String(input.search));
        params.set('limit', String(input.limit ?? 50));
        params.set('offset', String(input.offset ?? 0));
        const res = await createClient().get(`/erp/products?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'erp_get_product',
    description: 'Get a single ERP product by ID.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        product_id: { type: 'string' },
      },
      required: ['product_id'],
    },
    handler: async (input: Record<string, unknown>) => {
      try {
        const { product_id } = z.object({ product_id: z.string() }).parse(input);
        const res = await createClient().get(`/erp/products/${product_id}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'erp_list_suppliers',
    description: 'List ERP suppliers.',
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
        const res = await createClient().get(`/erp/suppliers?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'erp_list_purchase_orders',
    description: 'List ERP purchase orders.',
    inputSchema: {
      type: 'object' as const,
      properties: {
        status: { type: 'string', description: 'draft | sent | approved | received | cancelled' },
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
        const res = await createClient().get(`/erp/purchase-orders?${params.toString()}`);
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
  {
    name: 'erp_inventory_report',
    description: 'Get current inventory levels and stock report.',
    inputSchema: {
      type: 'object' as const,
      properties: {},
      required: [],
    },
    handler: async (_input: Record<string, unknown>) => {
      try {
        const res = await createClient().get('/erp/inventory/report');
        return { content: [{ type: 'text', text: JSON.stringify(res.data, null, 2) }] };
      } catch (err) {
        return { content: [{ type: 'text', text: `Error: ${apiError(err)}` }], isError: true };
      }
    },
  },
];
