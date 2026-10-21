import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { DueDatePicker } from './DueDatePicker';

import { I18nextProvider } from 'react-i18next';
import { createInstance } from 'i18next';
import enCommon from '../../locales/en/flowCommon.json';
import enFlow from '../../locales/en/flow.json';

const i18n = createInstance();
void i18n.init({
  lng: 'en',
  resources: { en: { translation: { flowCommon: enCommon, flow: enFlow } } },
  initAsync: false,
});
const renderPicker = (element: React.ReactNode) =>
  render(<I18nextProvider i18n={i18n}>{element}</I18nextProvider>);

describe('DueDatePicker clear control', () => {
  it('renders the clear control as a real button with an aria-label', () => {
    renderPicker(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const clearButton = screen.getByRole('button', { name: enCommon.datePicker.clearDate });
    expect(clearButton.tagName).toBe('BUTTON');
    expect(clearButton).toHaveAttribute('type', 'button');
  });

  it('is not nested inside the toggle button', () => {
    renderPicker(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const [toggleButton, clearButton] = screen.getAllByRole('button');
    expect(clearButton).toBe(screen.getByRole('button', { name: enCommon.datePicker.clearDate }));
    expect(clearButton).not.toBe(toggleButton);
    expect(toggleButton.contains(clearButton)).toBe(false);
  });

  it('is reachable via Tab and operable via keyboard alone', async () => {
    const user = userEvent.setup();
    const handleChange = vi.fn();
    renderPicker(<DueDatePicker value="2026-08-10" onChange={handleChange} />);

    const [toggleButton] = screen.getAllByRole('button');
    await user.tab();
    expect(toggleButton).toHaveFocus();

    await user.tab();
    const clearButton = screen.getByRole('button', { name: enCommon.datePicker.clearDate });
    expect(clearButton).toHaveFocus();

    await user.keyboard('{Enter}');
    expect(handleChange).toHaveBeenCalledWith('');
  });

  it('does not render the clear control when there is no value', () => {
    renderPicker(<DueDatePicker value="" onChange={vi.fn()} />);

    expect(
      screen.queryByRole('button', { name: enCommon.datePicker.clearDate })
    ).not.toBeInTheDocument();
  });

  it('does not render the clear control when disabled', () => {
    renderPicker(<DueDatePicker value="2026-08-10" onChange={vi.fn()} disabled />);

    expect(
      screen.queryByRole('button', { name: enCommon.datePicker.clearDate })
    ).not.toBeInTheDocument();
  });

  it('closes with Escape and restores focus to the trigger', async () => {
    const user = userEvent.setup();
    renderPicker(<DueDatePicker onChange={vi.fn()} />);
    const trigger = screen.getByRole('button', { name: enCommon.datePicker.setDate });
    await user.click(trigger);
    expect(screen.getByRole('dialog')).toBeVisible();
    await user.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('selects a shortcut, closes the picker and returns focus', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderPicker(<DueDatePicker onChange={onChange} />);
    const trigger = screen.getByRole('button', { name: enCommon.datePicker.setDate });
    await user.click(trigger);
    await user.click(screen.getByRole('button', { name: enCommon.datePicker.today, exact: true }));
    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    expect(onChange).toHaveBeenCalledWith(today);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('dismisses on an outside click', async () => {
    const user = userEvent.setup();
    renderPicker(
      <>
        <DueDatePicker onChange={vi.fn()} />
        <button>Outside</button>
      </>
    );
    await user.click(screen.getByRole('button', { name: enCommon.datePicker.setDate }));
    await user.click(screen.getByRole('button', { name: 'Outside' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('opens the date picker when clicking anywhere on the toggle button, not just the text node', async () => {
    const user = userEvent.setup();
    renderPicker(<DueDatePicker value="2026-08-10" onChange={vi.fn()} />);

    const [toggleButton] = screen.getAllByRole('button');
    await user.click(toggleButton);

    expect(screen.getByText(enCommon.datePicker.today)).toBeInTheDocument();
  });
});
