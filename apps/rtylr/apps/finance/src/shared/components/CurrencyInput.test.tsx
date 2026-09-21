import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import React from 'react';
import CurrencyInput from './CurrencyInput';

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
