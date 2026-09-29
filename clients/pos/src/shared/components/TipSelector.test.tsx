/**
 * The custom tip was read with parseFloat. parseFloat stops at the first
 * character it cannot parse, so a French cashier typing a 12,50 tip produced
 * 12: the customer kept all but 12 of the change and the order recorded the
 * rest as the amount paid. These cases assert the exact tip the parent
 * receives, because that number is what ends up on the payment request.
 *
 * The input is still `type="number"`, so the browser's own value sanitisation
 * runs first and can empty the field before a handler sees it. That is the
 * paired `type="text"` follow-up, not this fix; what is asserted here is that
 * whatever text does arrive is read at its true value.
 */
import React from 'react';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

let currentLanguage = 'en';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    i18n: { resolvedLanguage: currentLanguage, language: currentLanguage },
  }),
}));

vi.mock('../utils/format', () => ({ formatCurrency: (n: number) => `$${n.toFixed(2)}` }));

vi.mock('../stores/settingsStore', () => ({
  useSettingsStore: (selector: (state: { tipPresets: number[] }) => unknown) =>
    selector({ tipPresets: [0.1, 0.15, 0.18, 0.2] }),
}));

import TipSelector from './TipSelector';

const CASES: [locale: string, typed: string, expectedTip: number][] = [
  // The locale's own convention.
  ['fr', '12,50', 12.5],
  ['de', '12,50', 12.5],
  ['es', '12,50', 12.5],
  ['it', '12,50', 12.5],
  ['en', '12.50', 12.5],
  ['ar', '12.50', 12.5],
  // The other convention still reads at its true value.
  ['fr', '12.50', 12.5],
  ['en', '12,50', 12.5],
  // Grouping, including the narrow space a French copy carries.
  ['fr', '1\u202F234,56', 1234.56],
  ['en', '1 234.56', 1234.56],
  // Fractions, whole amounts, and a separator with nothing after it yet.
  ['fr', '0,5', 0.5],
  ['en', '0.5', 0.5],
  ['en', '12', 12],
  ['fr', '12,', 12],
  ['en', '12.', 12],
  // A tip cannot be negative.
  ['en', '-12.50', 0],
];

function typeTip(locale: string, typed: string): { tips: number[]; shown: string } {
  currentLanguage = locale;
  const tips: number[] = [];
  render(<TipSelector subtotal={100} onTipChange={(tip) => tips.push(tip)} />);
  const input = screen.getByPlaceholderText('0.00') as HTMLInputElement;
  fireEvent.change(input, { target: { value: typed } });
  return { tips, shown: input.value };
}

/** The tip the parent is left holding, whether or not a change was emitted. */
function effectiveTip(tips: number[]): number {
  return tips.length > 0 ? tips[tips.length - 1] : 0;
}

describe('TipSelector custom tip', () => {
  afterEach(cleanup);

  it.each(CASES)('reads %s "%s" as a tip of %d', (locale, typed, expectedTip) => {
    expect(effectiveTip(typeTip(locale, typed).tips)).toBe(expectedTip);
  });

  it('does not under-charge a French tip by a factor of a hundred', () => {
    // parseFloat('12,50') is 12, and that is what reached the order before.
    expect(Number.parseFloat('12,50')).toBe(12);
    const tip = effectiveTip(typeTip('fr', '12,50').tips);
    expect(tip).toBe(12.5);
    expect(tip).not.toBe(12);
  });

  it('does not over-charge by stripping the separator out instead', () => {
    // digitsOnly('12,50') is '1250': the mirror image of the same mistake.
    expect('12,50'.replace(/\D/g, '')).toBe('1250');
    expect(effectiveTip(typeTip('fr', '12,50').tips)).not.toBe(1250);
  });

  it('never turns text that is not an amount into money', () => {
    // Rejected outright by the number sanitiser, so the field empties and no
    // tip is emitted; either way the tip is zero and never a wrong amount.
    for (const [locale, typed] of [
      ['fr', 'abc'],
      ['en', ''],
      ['fr', 'free'],
    ] as [string, string][]) {
      cleanup();
      expect(effectiveTip(typeTip(locale, typed).tips)).toBe(0);
    }
  });

  it('leaves the operator looking at what they typed', () => {
    expect(typeTip('fr', '12,50').shown).toBe('12,50');
  });

  it('keeps a decimal separator usable while it is being typed', () => {
    currentLanguage = 'fr';
    const tips: number[] = [];
    render(<TipSelector subtotal={100} onTipChange={(tip) => tips.push(tip)} />);
    const input = screen.getByPlaceholderText('0.00') as HTMLInputElement;

    fireEvent.change(input, { target: { value: '1' } });
    fireEvent.change(input, { target: { value: '12,' } });
    fireEvent.change(input, { target: { value: '12,50' } });

    // Reading the intermediate "12," back as 12 is harmless; reading "1,50" as
    // 1 is what silently under-charged every French sale.
    expect(tips).toEqual([1, 12, 12.5]);
    expect(input.value).toBe('12,50');
  });

  it('still applies a preset percentage of the subtotal', () => {
    currentLanguage = 'en';
    const tips: number[] = [];
    render(<TipSelector subtotal={100} onTipChange={(tip) => tips.push(tip)} />);

    fireEvent.click(screen.getByText('15%'));

    expect(effectiveTip(tips)).toBe(15);
  });
});
