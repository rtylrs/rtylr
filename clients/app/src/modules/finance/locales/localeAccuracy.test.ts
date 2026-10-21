import { describe, expect, it } from 'vitest';
import enInvoices from './en/financeInvoices.json';
import arInvoices from './ar/financeInvoices.json';
import frInvoices from './fr/financeInvoices.json';

describe('Finance invoice translations', () => {
  it('does not fall back to English invoice notifications in Arabic or French', () => {
    for (const key of Object.keys(enInvoices) as Array<keyof typeof enInvoices>) {
      expect(arInvoices[key]).not.toBe(enInvoices[key]);
      expect(frInvoices[key]).not.toBe(enInvoices[key]);
    }
  });
});
