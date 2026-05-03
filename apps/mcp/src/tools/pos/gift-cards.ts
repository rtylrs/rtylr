import type { Tool } from '../types.js';
import { ok, apiErr } from '../types.js';

export const getGiftCard: Tool = {
  name: 'pos_get_gift_card',
  description: 'Get a gift card by code, including balance and status.',
  inputSchema: {
    type: 'object',
    properties: {
      code: { type: 'string', description: 'Gift card code' },
    },
    required: ['code'],
  },
  handler: async (input, api) => {
    try {
      const res = await api.get(`/pos/gift-cards/${input.code}`);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};

export const redeemGiftCard: Tool = {
  name: 'pos_redeem_gift_card',
  description: 'Redeem a gift card against an order.',
  inputSchema: {
    type: 'object',
    properties: {
      code: { type: 'string', description: 'Gift card code' },
      amount: { type: 'string', description: 'Amount to redeem in cents' },
      order_id: { type: 'string', description: 'Order ID to apply the redemption to' },
    },
    required: ['code', 'amount', 'order_id'],
  },
  handler: async (input, api) => {
    try {
      const { code, ...body } = input;
      const res = await api.post(`/pos/gift-cards/${code}/redeem`, body);
      return ok(res.data);
    } catch (e) {
      return apiErr(e);
    }
  },
};
