import type { DiscountOrderRequest } from '@voxire/voxie/pos/orders/discounts/v1/types';

export type DiscountMode = 'percent' | 'flat';

interface BuildDiscountRequestParams {
  orderId: string | undefined;
  mode: DiscountMode;
  valueInput: string;
}

interface DiscountRequest {
  payload: DiscountOrderRequest;
}

function positiveInteger(value: unknown): number | null {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return null;
  return Math.floor(numeric);
}

function discountValue(mode: DiscountMode, valueInput: string): number {
  const value = Number(valueInput);
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error('Enter a discount greater than zero.');
  }
  return mode === 'percent' ? Math.min(100, value) : value;
}

export function buildDiscountRequest({
  orderId,
  mode,
  valueInput,
}: BuildDiscountRequestParams): DiscountRequest {
  const normalizedOrderId = positiveInteger(orderId);
  if (!normalizedOrderId) {
    throw new Error('Select a saved order before applying a discount.');
  }

  return {
    payload: {
      orderId: String(normalizedOrderId),
      discountType: mode,
      discountValue: discountValue(mode, valueInput),
      reason: '',
    },
  };
}
