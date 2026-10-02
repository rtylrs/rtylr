/**
 * Money formatting for amounts that arrive as major units.
 *
 * POS order money is not minor units. `pos_orders.subtotal`, `discount_amount`,
 * `tax_amount`, `total`, `amount_paid` and `pos_order_items.unit_price` are all
 * `DECIMAL(10,2)` in onion's `pos/schema.sql`, read out through
 * `sqlutil.DecimalToFloat`, and declared `float64` in
 * `pos/orders/v1/models.onk`. Nothing in that chain scales by 100, so a $12.50
 * order arrives as `12.5`.
 *
 * That is the opposite of the billing amounts, which really are minor units and
 * are named for it (`pricePerSeatCents`, `amountCents`). Those keep using
 * `centsToDisplay`; this is for everything that does not.
 *
 * Currency is a parameter rather than a hardcoded symbol. `Intl.NumberFormat`
 * takes the fraction digits from the currency, so a zero-decimal currency such
 * as JPY formats with no decimals and a three-decimal one such as KWD with
 * three, without this module knowing which is which.
 */

/**
 * Currency to assume when the payload carries none.
 *
 * The POS order contract has no currency field, so there is nothing to read per
 * order yet. Both places the backend stores one, `businesses.currency` and
 * `pos_locations.currency`, are `VARCHAR(10) DEFAULT 'USD'`, so USD is the
 * documented default rather than an invention here. Once the Order message
 * carries a currency, pass it through and this constant becomes the fallback
 * for a missing value only.
 */
export const DEFAULT_CURRENCY = 'USD';

/** Formats a major-unit amount, letting the currency decide the decimals. */
export function formatMoney(
  amount: number,
  currency: string = DEFAULT_CURRENCY,
  locale = 'en-US'
): string {
  const safe = Number.isFinite(amount) ? amount : 0;
  try {
    return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(safe);
  } catch {
    // An unknown or malformed currency code makes Intl throw. Falling back to
    // the code plus the number keeps the amount readable and keeps the code
    // visible, which is more useful than a crashed row or a wrong symbol.
    return `${currency} ${safe.toFixed(2)}`;
  }
}
