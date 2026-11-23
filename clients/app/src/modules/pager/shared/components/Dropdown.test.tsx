import React, { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { Dropdown } from './Dropdown';

const SIZES = [
  { value: 'small', label: 'Small' },
  { value: 'medium', label: 'Medium' },
  { value: 'large', label: 'Large' },
] as const;
type Size = (typeof SIZES)[number]['value'];

function Field({
  onChange = vi.fn(),
  disabled = false,
}: {
  onChange?: (size: Size) => void;
  disabled?: boolean;
}) {
  const [size, setSize] = useState<Size>('medium');
  return (
    <>
      <label id="size-label" htmlFor="size">
        Cup size
      </label>
      <Dropdown
        id="size"
        labelId="size-label"
        value={size}
        options={SIZES}
        disabled={disabled}
        onChange={(next) => {
          setSize(next);
          onChange(next);
        }}
      />
      <button type="button">Elsewhere</button>
    </>
  );
}

describe('Dropdown', () => {
  it('shows the chosen option and picks another with a click', () => {
    const onChange = vi.fn();
    render(<Field onChange={onChange} />);
    const field = screen.getByRole('combobox', { name: 'Cup size' });
    expect(field).toHaveTextContent('Medium');
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

    fireEvent.click(field);
    expect(field).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('option', { name: 'Medium' })).toHaveAttribute('aria-selected', 'true');
    fireEvent.click(screen.getByRole('option', { name: 'Large' }));

    expect(onChange).toHaveBeenCalledWith('large');
    expect(field).toHaveTextContent('Large');
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('works from the keyboard, starting on the chosen option', () => {
    const onChange = vi.fn();
    render(<Field onChange={onChange} />);
    const field = screen.getByRole('combobox', { name: 'Cup size' });

    fireEvent.keyDown(field, { key: 'ArrowDown' });
    expect(screen.getByRole('listbox')).toBeInTheDocument();
    const active = (): string | null =>
      document.getElementById(field.getAttribute('aria-activedescendant') ?? '')?.textContent ??
      null;
    expect(active()).toBe('Medium');

    fireEvent.keyDown(field, { key: 'ArrowUp' });
    expect(active()).toBe('Small');
    fireEvent.keyDown(field, { key: 'End' });
    expect(active()).toBe('Large');
    fireEvent.keyDown(field, { key: 'Enter' });

    expect(onChange).toHaveBeenCalledWith('large');
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });

  it('closes without changing on Escape or a click elsewhere', () => {
    const onChange = vi.fn();
    render(<Field onChange={onChange} />);
    const field = screen.getByRole('combobox', { name: 'Cup size' });

    fireEvent.click(field);
    fireEvent.keyDown(field, { key: 'ArrowDown' });
    fireEvent.keyDown(field, { key: 'Escape' });
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();

    fireEvent.click(field);
    fireEvent.pointerDown(screen.getByRole('button', { name: 'Elsewhere' }));
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
    expect(field).toHaveTextContent('Medium');
  });

  it('does not open while disabled', () => {
    render(<Field disabled />);
    const field = screen.getByRole('combobox', { name: 'Cup size' });
    expect(field).toBeDisabled();
    fireEvent.click(field);
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
  });
});
