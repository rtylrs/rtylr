import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DatePicker } from '../index';

describe('DatePicker', () => {
  it('shows a placeholder when there is no value', () => {
    render(<DatePicker onChange={vi.fn()} labels={{ placeholder: 'Select date' }} />);
    expect(screen.getByRole('button')).toHaveTextContent('Select date');
  });

  it('shows the formatted value when one is set', () => {
    render(<DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />);
    expect(screen.getByRole('button')).toHaveTextContent('Aug 16, 2026');
  });

  it('opens a calendar panel on click and closes it after picking a date', () => {
    const handleChange = vi.fn();
    render(<DatePicker value="2026-08-16" onChange={handleChange} locale="en-US" />);

    fireEvent.click(screen.getByRole('button', { name: /aug 16, 2026/i }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('gridcell', { name: /august 20, 2026/i }));
    expect(handleChange).toHaveBeenCalledWith('2026-08-20');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('closes the panel on Escape without changing the value', () => {
    const handleChange = vi.fn();
    render(<DatePicker value="2026-08-16" onChange={handleChange} locale="en-US" />);
    fireEvent.click(screen.getByRole('button'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(handleChange).not.toHaveBeenCalled();
  });

  it('closes the panel on an outside click', () => {
    render(
      <div>
        <DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />
        <button>outside</button>
      </div>
    );
    fireEvent.click(screen.getByRole('button', { name: /aug 16, 2026/i }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByRole('button', { name: 'outside' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('shows a clear control only when clearable and a value is set, and it does not open the panel', () => {
    const handleChange = vi.fn();
    render(
      <DatePicker
        value="2026-08-16"
        onChange={handleChange}
        locale="en-US"
        clearable
        labels={{ clear: 'Clear date' }}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Clear date' }));
    expect(handleChange).toHaveBeenCalledWith('');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('does not open when disabled', () => {
    render(<DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" disabled />);
    fireEvent.click(screen.getByRole('button'));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('is reachable via a label using its id, like a native form control', () => {
    render(
      <div>
        <label htmlFor="due-date">Due date</label>
        <DatePicker id="due-date" value="2026-08-16" onChange={vi.fn()} locale="en-US" />
      </div>
    );
    expect(screen.getByLabelText('Due date')).toHaveAttribute('data-value', '2026-08-16');
  });
});
