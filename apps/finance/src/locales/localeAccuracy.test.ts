import { describe, expect, it } from 'vitest';
import enInvoices from './en/invoices.json';
import arInvoices from './ar/invoices.json';
import frInvoices from './fr/invoices.json';

describe('Finance invoice translations', () => {
  it('does not fall back to English invoice notifications in Arabic or French', () => {
    for (const key of Object.keys(enInvoices) as Array<keyof typeof enInvoices>) {
      expect(arInvoices[key]).not.toBe(enInvoices[key]);
      expect(frInvoices[key]).not.toBe(enInvoices[key]);
    }
  });
});
