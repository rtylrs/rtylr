export interface TaxBreakdownInput {
  subtotal: number;
  taxRate: number;
  inclusive?: boolean;
  discount?: number;
}

export interface TaxBreakdownResult {
  subtotal: number;
  discount: number;
  taxableAmount: number;
  taxAmount: number;
  total: number;
}

export function computeTaxBreakdown({
  subtotal,
  taxRate,
  inclusive = false,
  discount = 0,
}: TaxBreakdownInput): TaxBreakdownResult {
  const rate = Math.max(0, taxRate) / 100;
  const discountAmount = Math.max(0, Math.min(discount, subtotal));
  const afterDiscount = subtotal - discountAmount;

  if (inclusive) {
    const taxAmount = afterDiscount - afterDiscount / (1 + rate);
    return {
      subtotal,
      discount: discountAmount,
      taxableAmount: afterDiscount / (1 + rate),
      taxAmount,
      total: afterDiscount,
    };
  }

  const taxAmount = afterDiscount * rate;
  return {
    subtotal,
    discount: discountAmount,
    taxableAmount: afterDiscount,
    taxAmount,
    total: afterDiscount + taxAmount,
  };
}

export function formatTaxRate(rate: number, fractionDigits = 2): string {
  return `${rate.toFixed(fractionDigits)}%`;
}

export function roundMoney(value: number, decimals = 2): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}
