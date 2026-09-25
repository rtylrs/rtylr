import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useCaretSettingsStore } from '../../stores/settingsStore';
import ModePicker from './ModePicker';

describe('ModePicker', () => {
  beforeEach(() => {
    useCaretSettingsStore.setState({ mode: 'analyst' });
  });

  it('uses a compact single-line menu with a selected state', () => {
    render(<ModePicker compact />);

    const trigger = screen.getByRole('button', { name: 'Mode: Analyst' });
    fireEvent.click(trigger);

    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    expect(screen.getAllByRole('menuitemradio')).toHaveLength(4);
    expect(
      screen.getByRole('menuitemradio', { name: 'Analyst' }).getAttribute('aria-checked')
    ).toBe('true');
    expect(screen.queryByText('Reports, tables, and KPI breakdowns. Read-only.')).toBeNull();

    fireEvent.click(screen.getByRole('menuitemradio', { name: 'Ask' }));

    expect(useCaretSettingsStore.getState().mode).toBe('ask');
    expect(screen.queryByRole('menu')).toBeNull();
  });

  it('consumes Escape so closing the menu does not close the Caret panel', () => {
    const onWindowKeyDown = vi.fn();
    window.addEventListener('keydown', onWindowKeyDown);
    render(<ModePicker compact />);

    const trigger = screen.getByRole('button', { name: 'Mode: Analyst' });
    fireEvent.click(trigger);
    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByRole('menu')).toBeNull();
    expect(document.activeElement).toBe(trigger);
    expect(onWindowKeyDown).not.toHaveBeenCalled();
    window.removeEventListener('keydown', onWindowKeyDown);
  });
});
