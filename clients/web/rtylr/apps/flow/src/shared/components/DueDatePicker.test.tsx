import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { DueDatePicker } from './DueDatePicker';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'en-US' },
  }),
}));

describe('DueDatePicker clear control', () => {
  it('renders the clear control as a real button with an aria-label', () => {
    render(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const clearButton = screen.getByRole('button', { name: 'common.datePicker.clearDate' });
    expect(clearButton.tagName).toBe('BUTTON');
    expect(clearButton).toHaveAttribute('type', 'button');
  });

  it('is not nested inside the toggle button', () => {
    render(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const [toggleButton, clearButton] = screen.getAllByRole('button');
    expect(clearButton).toBe(screen.getByRole('button', { name: 'common.datePicker.clearDate' }));
    expect(clearButton).not.toBe(toggleButton);
    expect(toggleButton.contains(clearButton)).toBe(false);
  });

  it('is reachable via Tab and operable via keyboard alone', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    render(<DueDatePicker value="2026-08-10" onChange={handleChange} />);

    const [toggleButton] = screen.getAllByRole('button');
    await user.tab();
    expect(toggleButton).toHaveFocus();

    await user.tab();
    const clearButton = screen.getByRole('button', { name: 'common.datePicker.clearDate' });
    expect(clearButton).toHaveFocus();

    await user.keyboard('{Enter}');
    expect(handleChange).toHaveBeenCalledWith('');
  });

  it('does not render the clear control when there is no value', () => {
    render(<DueDatePicker value="" onChange={vi.fn()} />);

    expect(
      screen.queryByRole('button', { name: 'common.datePicker.clearDate' })
    ).not.toBeInTheDocument();
  });

  it('does not render the clear control when disabled', () => {
    render(<DueDatePicker value="2026-08-10" onChange={vi.fn()} disabled />);

    expect(
      screen.queryByRole('button', { name: 'common.datePicker.clearDate' })
    ).not.toBeInTheDocument();
  });

  it('gives the toggle button its own padding so the pill click area is not a dead zone on the wrapper', () => {
    render(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const [toggleButton] = screen.getAllByRole('button');
    expect(toggleButton.className).toMatch(/\bpx-2\b/);
    expect(toggleButton.className).toMatch(/\bpy-1\b/);

    const wrapper = toggleButton.parentElement;
    expect(wrapper?.className).not.toMatch(/\bpx-2\b/);
    expect(wrapper?.className).not.toMatch(/\bpy-1\b/);
  });

  it('opens the date picker when clicking anywhere on the toggle button, not just the text node', async () => {
    const user = userEvent.setup();
    render(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const [toggleButton] = screen.getAllByRole('button');
    await user.click(toggleButton);

    expect(screen.getByText('common.datePicker.today')).toBeInTheDocument();
  });
});
