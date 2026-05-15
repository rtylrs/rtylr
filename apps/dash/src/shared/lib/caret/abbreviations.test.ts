import { describe, expect, it } from 'vitest';
import { expandAbbreviations } from './abbreviations';

describe('expandAbbreviations', () => {
  it('uppercases known abbreviations', () => {
    expect(expandAbbreviations('check the pos and crm')).toBe('check the POS and CRM');
  });

  it('only matches whole words', () => {
    expect(expandAbbreviations('postal sms')).toBe('postal SMS');
  });

  it('preserves unknown words', () => {
    expect(expandAbbreviations('hello world')).toBe('hello world');
  });
});
