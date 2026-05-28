// Pure validators for payment inputs.

/** Returns true if the string represents a valid positive cash amount. */
export function isValidCashAmount(input: string): boolean {
  const num = Number(input.trim());
  return Number.isFinite(num) && num > 0;
}

/** Returns true if the string is a non-empty reason (used for voids/refunds). */
export function isValidReason(reason: string): boolean {
  return reason.trim().length > 0;
}

/** Returns true if a discount percentage is within the 0-100 range. */
export function isValidDiscountPercent(pct: number): boolean {
  return Number.isFinite(pct) && pct > 0 && pct <= 100;
}

/** Returns true if a flat discount amount is positive and finite. */
export function isValidFlatDiscount(amount: number): boolean {
  return Number.isFinite(amount) && amount > 0;
}

/** Returns true if a tip percentage is between 0 and 100 inclusive. */
export function isValidTipPercent(pct: number): boolean {
  return Number.isFinite(pct) && pct >= 0 && pct <= 100;
}

/** Returns true if a tip amount is non-negative and finite. */
export function isValidTipAmount(amount: number): boolean {
  return Number.isFinite(amount) && amount >= 0;
}

/** Returns true if a gift card code looks structurally valid (non-empty alphanumeric). */
export function isValidGiftCardCode(code: string): boolean {
  return /^[A-Za-z0-9-]{4,32}$/.test(code.trim());
}

/** Returns true if a PIN is all digits and between 4-6 digits long. */
export function isValidPin(pin: string): boolean {
  return /^\d{4,6}$/.test(pin.trim());
}

/** Returns true if the number of split seats is at least 2 and an integer. */
export function isValidSplitSeats(seats: number): boolean {
  return Number.isInteger(seats) && seats >= 2;
}
