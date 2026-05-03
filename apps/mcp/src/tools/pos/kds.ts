import type { Tool } from '../types.js';
import { ok, apiErr, qs } from '../types.js';

export const kdsListOrders: Tool = {
  name: 'pos_kds_list_orders',
  description: 'List orders on the Kitchen Display System (KDS), optionally filtered by station and status.',
  inputSchema: {
    type: 'object',
    properties: {
      station_id: { type: 'string', description: 'Filter by KDS station ID' },
      status: { type: 'string', description: 'Filter by status', enum: ['pending', 'preparing'] },
      limit: { type: 'string' },
    },
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/kds/orders${qs(input, ['station_id', 'status', 'limit'])}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
