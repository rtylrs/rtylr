import type { DiscountOrderRequest } from '@voxire/voxie/pos/orders/discounts/v1/types';

export type DiscountMode = 'percent' | 'flat';

interface BuildDiscountRequestParams {
  orderId: string | undefined;
  mode: DiscountMode;
  valueInput: string;
  reason?: string;
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
  // A typo like "150" must be rejected outright, never silently turned into
  // a free order.
  if (mode === 'percent' && value > 100) {
    throw new Error('Percent discounts cannot exceed 100%.');
  }
  return value;
}

export function buildDiscountRequest({
  orderId,
  mode,
  valueInput,
  reason = '',
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
      reason: reason.trim(),
    },
  };
}
