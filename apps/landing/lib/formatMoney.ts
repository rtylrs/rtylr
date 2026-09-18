/**
 * $12.50 in English and Arabic, 12,50 $ in French - the convention
 * home.features.demos.pos.chargeAmount's own translated string already
 * establishes. Centralised so the demos that display a formatted dollar
 * amount in code (not through a translated string) can't drift into a
 * different placement of the symbol, which is exactly what happened
 * before this existed: four call sites each hardcoded "$" as a prefix.
 */
export function formatMoney(amount: number, locale: string, fractionDigits = 2): string {
  const n = amount.toLocaleString(locale, {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  });
  return locale === 'fr' ? `${n} $` : `$${n}`;
}

/** Same convention for a compact "thousands" figure, e.g. $42.8k / 42,8 k$. */
export function formatMoneyK(amountInThousands: number, locale: string): string {
  const n = amountInThousands.toLocaleString(locale, {
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
  });
  return locale === 'fr' ? `${n} k$` : `$${n}k`;
}
