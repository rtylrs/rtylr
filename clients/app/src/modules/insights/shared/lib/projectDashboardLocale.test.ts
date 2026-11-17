import { describe, expect, it } from 'vitest';
import en from '../../../../locales/en/insights.json';
import fr from '../../../../locales/fr/insights.json';
import ar from '../../../../locales/ar/insights.json';

type Copy = Record<string, string>;

const PLURAL = /_(zero|one|two|few|many|other)$/;
const base = (key: string): string => key.replace(PLURAL, '');
const bases = (copy: Copy): string[] => [...new Set(Object.keys(copy).map(base))].sort();

const enCopy = en.projectDashboard as Copy;
const frCopy = fr.projectDashboard as Copy;
const arCopy = ar.projectDashboard as Copy;

// Counted strings, by the plural forms each language needs.
const counted = bases(enCopy).filter((key) => `${key}_other` in enCopy);

describe('project dashboard copy (T-567)', () => {
  it('has every English string in French and Arabic', () => {
    expect(bases(frCopy)).toEqual(bases(enCopy));
    expect(bases(arCopy)).toEqual(bases(enCopy));
  });

  it('is translated, not copied from English', () => {
    for (const key of ['title', 'notFoundDescription', 'noInvoices', 'nothingOverdue']) {
      expect(frCopy[key], `fr ${key}`).not.toBe(enCopy[key]);
      expect(arCopy[key], `ar ${key}`).not.toBe(enCopy[key]);
    }
  });

  it.each(counted)('gives %s every plural form', (key) => {
    for (const form of ['one', 'other']) expect(enCopy[`${key}_${form}`]).toBeTruthy();
    for (const form of ['one', 'many', 'other']) expect(frCopy[`${key}_${form}`]).toBeTruthy();
    for (const form of ['zero', 'one', 'two', 'few', 'many', 'other']) {
      expect(arCopy[`${key}_${form}`], `ar ${key}_${form}`).toBeTruthy();
    }
  });
});
