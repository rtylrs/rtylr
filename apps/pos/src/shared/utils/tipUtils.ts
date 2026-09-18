// Pure helpers for tip calculation used at checkout.

/** Suggested tip percentages shown in the UI. */
export const SUGGESTED_TIP_PCTS = [10, 15, 18, 20, 25] as const;

export type SuggestedTipPct = (typeof SUGGESTED_TIP_PCTS)[number];

/** Calculate a tip amount from a subtotal and a percentage (0-100). */
export function calcTipAmount(subtotal: number, pct: number): number {
  if (!Number.isFinite(subtotal) || subtotal < 0) return 0;
  if (!Number.isFinite(pct) || pct < 0) return 0;
  return Math.round(subtotal * (pct / 100) * 100) / 100;
}

/** Calculate the effective tip percentage from an amount and a subtotal. */
export function calcTipPercent(tipAmount: number, subtotal: number): number {
  if (!Number.isFinite(tipAmount) || tipAmount <= 0) return 0;
  if (!Number.isFinite(subtotal) || subtotal <= 0) return 0;
  return Math.round((tipAmount / subtotal) * 100 * 100) / 100;
}

/** Returns true if the tip amount is reasonable (less than 100% of the subtotal). */
export function isTipReasonable(tipAmount: number, subtotal: number): boolean {
  if (!Number.isFinite(tipAmount) || tipAmount < 0) return false;
  if (!Number.isFinite(subtotal) || subtotal <= 0) return tipAmount === 0;
  return tipAmount <= subtotal;
}

/** Return all suggested tip amounts for a given subtotal. */
export function suggestedTipAmounts(subtotal: number): Record<SuggestedTipPct, number> {
  return SUGGESTED_TIP_PCTS.reduce(
    (acc, pct) => {
      acc[pct] = calcTipAmount(subtotal, pct);
      return acc;
    },
    {} as Record<SuggestedTipPct, number>,
  );
}
