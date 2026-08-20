import React from 'react';
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Calendar, toISODate, todayISODate } from '../index';

describe('Calendar', () => {
  it('marks the passed-in value as selected', () => {
    render(<Calendar value="2026-08-16" onChange={vi.fn()} />);
    const selected = screen.getByRole('gridcell', { selected: true });
    expect(selected).toHaveTextContent('16');
  });

  it('calls onChange with the ISO date of the clicked day', () => {
    const handleChange = vi.fn();
    render(<Calendar value="2026-08-16" onChange={handleChange} />);
    fireEvent.click(screen.getByRole('gridcell', { name: /august 20, 2026/i }));
    expect(handleChange).toHaveBeenCalledWith('2026-08-20');
  });

  it('disables days before minDate and does not fire onChange for them', () => {
    const handleChange = vi.fn();
    render(<Calendar value="2026-08-16" onChange={handleChange} minDate="2026-08-10" />);
    const tooEarly = screen.getByRole('gridcell', { name: /august 5, 2026/i });
    expect(tooEarly).toBeDisabled();
    fireEvent.click(tooEarly);
    expect(handleChange).not.toHaveBeenCalled();
  });

  it('disables days after maxDate', () => {
    render(<Calendar value="2026-08-16" onChange={vi.fn()} maxDate="2026-08-20" />);
    expect(screen.getByRole('gridcell', { name: /august 25, 2026/i })).toBeDisabled();
    expect(screen.getByRole('gridcell', { name: /august 18, 2026/i })).not.toBeDisabled();
  });

  it('moves to the next and previous month', () => {
    render(<Calendar value="2026-08-16" onChange={vi.fn()} />);
    expect(screen.getByText('August 2026')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Next month' }));
    expect(screen.getByText('September 2026')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Previous month' }));
    fireEvent.click(screen.getByRole('button', { name: 'Previous month' }));
    expect(screen.getByText('July 2026')).toBeInTheDocument();
  });

  it('moves focus one day at a time with the arrow keys, crossing month boundaries', () => {
    render(<Calendar value="2026-08-31" onChange={vi.fn()} />);
    const day31 = screen.getByRole('gridcell', { name: /august 31, 2026/i });
    day31.focus();
    fireEvent.keyDown(day31, { key: 'ArrowRight' });
    expect(screen.getByText('September 2026')).toBeInTheDocument();
  });

  it('jumps a full week with ArrowDown/ArrowUp', () => {
    render(<Calendar value="2026-08-16" onChange={vi.fn()} />);
    const day16 = screen.getByRole('gridcell', { name: /august 16, 2026/i });
    day16.focus();
    fireEvent.keyDown(day16, { key: 'ArrowDown' });
    expect(screen.getByRole('gridcell', { name: /august 23, 2026/i })).toHaveAttribute(
      'tabindex',
      '0'
    );
  });

  it('renders a Today shortcut only when a label is supplied for it, and it respects minDate', () => {
    const handleChange = vi.fn();
    const { rerender } = render(<Calendar value="2026-08-16" onChange={handleChange} />);
    expect(screen.queryByText('Today')).not.toBeInTheDocument();

    rerender(
      <Calendar value="2026-08-16" onChange={handleChange} labels={{ today: 'Today' }} />
    );
    fireEvent.click(screen.getByText('Today'));
    expect(handleChange).toHaveBeenCalledWith(todayISODate());
  });

  it('marks today with aria-current without marking it selected', () => {
    render(<Calendar value="2026-08-16" onChange={vi.fn()} />);
    const today = screen.getByRole('gridcell', {
      name: new RegExp(
        new Intl.DateTimeFormat(undefined, { dateStyle: 'full' }).format(new Date()),
        'i'
      ),
    });
    expect(today).toHaveAttribute('aria-current', 'date');
  });
});

describe('toISODate', () => {
  it('zero-pads month and day, and is 0-indexed on month like Date', () => {
    expect(toISODate(2026, 0, 5)).toBe('2026-01-05');
    expect(toISODate(2026, 11, 31)).toBe('2026-12-31');
  });
});
