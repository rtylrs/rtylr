import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeAll, describe, expect, it } from 'vitest';
import React from 'react';
import { i18next } from '@voxire/i18n';
import CurrencyInput from './CurrencyInput';
import { minorToMajor } from '../lib/money';
import { initTestI18n } from '../testing/i18n';

/** Mirrors the callers: the amount is parent state in minor units. */
const Harness: React.FC<{ onValue?: (value: number) => void }> = ({ onValue }) => {
  const [value, setValue] = React.useState(0);
  return (
    <CurrencyInput
      value={value}
      onChange={(next) => {
        setValue(next);
        onValue?.(next);
      }}
    />
  );
};

beforeAll(async () => {
  await initTestI18n('en');
});

describe('CurrencyInput', () => {
  it('keeps the decimal separator while typing, so "1.50" is 150 cents', async () => {
    const user = userEvent.setup();
    const changes: number[] = [];
    render(<Harness onValue={(value) => changes.push(value)} />);
    const input = screen.getByRole('textbox');

    await user.type(input, '1.50');

    expect(input).toHaveValue('1.50');
    expect(changes[changes.length - 1]).toBe(150);
  });

  it('normalises the displayed amount on blur', async () => {
    const user = userEvent.setup();
    render(<Harness />);
    const input = screen.getByRole('textbox');

    await user.type(input, '1.50');
    await user.tab();

    expect(input).toHaveValue('1.5');
  });

  it('reports whole amounts in minor units', async () => {
    const user = userEvent.setup();
    const changes: number[] = [];
    render(<Harness onValue={(value) => changes.push(value)} />);
    const input = screen.getByRole('textbox');

    await user.type(input, '12');

    expect(changes[changes.length - 1]).toBe(1200);
  });
});

/**
 * The sanitiser kept [0-9.-] only, so a French user typing "12,50" got 1250:
 * a hundred times the amount, and the form then submitted it as real money.
 * Every case below is read in the locale the operator is actually using.
 */
const CASES: [locale: string, typed: string, expectedMajor: number][] = [
  ['en', '12.50', 12.5],
  ['fr', '12,50', 12.5],
  ['ar', '12,50', 12.5],
  ['fr', '12.50', 12.5],
  ['fr', '12.5', 12.5],
  ['de', '12,50', 12.5],
  ['es', '12,50', 12.5],
  ['it', '12,50', 12.5],
  ['de', '1.234,56', 1234.56],
  ['en', '1,234.56', 1234.56],
  // Pasted amounts carry the narrow no-breaking space fr formats with.
  ['fr', '1\u202F234,56', 1234.56],
  ['en', '1 234.56', 1234.56],
  ['fr', '1 234,56', 1234.56],
  ['fr', '\u20AC1\u202F234,56', 1234.56],
  ['en', '12', 12],
  ['en', '12.', 12],
  ['fr', '12,', 12],
  ['en', '0.5', 0.5],
  ['fr', '0,5', 0.5],
  ['en', '-12.50', -12.5],
  ['fr', '\u221212,50', -12.5],
  // The two spellings both conventions share. The locale's own convention wins.
  ['en', '1,234', 1234],
  ['en', '1.234', 1.234],
  ['de', '1.234', 1234],
  ['de', '1,234', 1.234],
  ['en', '12,5', 12.5],
  ['fr', '12,5', 12.5],
  // Text that is not an amount must not become money.
  ['en', 'abc', 0],
  ['en', '1.2.3', 1.23],
];

async function typeIn(locale: string, text: string) {
  await i18next.changeLanguage(locale);
  const changes: number[] = [];
  render(<Harness onValue={(value) => changes.push(value)} />);
  const input = screen.getByRole('textbox') as HTMLInputElement;
  fireEvent.change(input, { target: { value: text } });
  return { minor: changes[changes.length - 1], shown: input.value };
}

describe('CurrencyInput reads an amount in the active locale', () => {
  afterEach(async () => {
    await i18next.changeLanguage('en');
  });

  it.each(CASES)('reads %s "%s" as %d major units', async (locale, typed, expectedMajor) => {
    const { minor } = await typeIn(locale, typed);
    expect(minor).toBe(Math.round(expectedMajor * 100));
  });

  it('leaves the typed text on screen so the operator can check it', async () => {
    const { shown } = await typeIn('fr', '12,50');
    expect(shown).toBe('12,50');
  });

  it('emits the true decimal amount, not a locale-naive parseFloat of the text', async () => {
    const { minor } = await typeIn('fr', '12,50');
    // Every caller turns minor units back into major units arithmetically and
    // sends that to the API. parseFloat('12,50') would be 12.
    expect(minor / 100).toBe(12.5);
    expect(Number.parseFloat('12,50')).not.toBe(minor / 100);
  });

  it('returns to zero when the amount is cleared', async () => {
    await i18next.changeLanguage('fr');
    const changes: number[] = [];
    render(<Harness onValue={(value) => changes.push(value)} />);
    const input = screen.getByRole('textbox');

    fireEvent.change(input, { target: { value: '12,50' } });
    fireEvent.change(input, { target: { value: '' } });

    expect(changes).toEqual([1250, 0]);
  });
});

/**
 * What NewExpense, Quotes, Bills and NewInvoice do on submit: hold the minor
 * units CurrencyInput emits, convert them arithmetically, and send the major
 * units. The conversion is where a wrong parse becomes a wrong request body.
 */
const SubmitHarness: React.FC<{ onSubmit: (amount: number) => void }> = ({ onSubmit }) => {
  const [amount, setAmount] = React.useState(0);
  return (
    <>
      <CurrencyInput value={amount} onChange={setAmount} />
      <button type="button" onClick={() => onSubmit(minorToMajor(amount, 'USD'))}>
        Save
      </button>
    </>
  );
};

describe('CurrencyInput submit path', () => {
  afterEach(async () => {
    await i18next.changeLanguage('en');
  });

  it.each([
    ['fr', '12,50', 12.5],
    ['de', '1.234,56', 1234.56],
    ['en', '12.50', 12.5],
    ['en', '-12.50', -12.5],
  ])('sends %s "%s" to the API as %d', async (locale, typed, expected) => {
    await i18next.changeLanguage(locale);
    const submitted: number[] = [];
    render(<SubmitHarness onSubmit={(amount) => submitted.push(amount)} />);

    fireEvent.change(screen.getByRole('textbox'), { target: { value: typed } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(submitted).toEqual([expected]);
  });
});
