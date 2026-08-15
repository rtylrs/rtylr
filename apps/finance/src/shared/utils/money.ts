// Re-export central money helpers from shared/lib/money to maintain backward compatibility.

export type { Currency } from '../lib/money';
export {
  getCurrencySymbol,
  formatMoney,
  minorToMajor,
  majorToMinor,
  fractionDigitsForCurrency,
} from '../lib/money';
