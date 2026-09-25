import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import ConfirmModal from './ConfirmModal';
import FilterDrawer from './FilterDrawer';
import KeyboardShortcutsModal from './KeyboardShortcutsModal';

describe('ERP transient surface accessibility', () => {
  it('gives confirmation dialogs an accessible name and named close control', () => {
    render(
      <ConfirmModal
        open
        title="Delete category"
        description="This cannot be undone."
        onConfirm={vi.fn()}
        onCancel={vi.fn()}
      />
    );

    expect(screen.getByRole('dialog', { name: 'Delete category' })).toHaveAttribute(
      'aria-describedby'
    );
    expect(screen.getByRole('button', { name: 'Close dialog' })).toBeInTheDocument();
  });

  it('names keyboard and filter surfaces and closes them on Escape', () => {
    const onKeyboardClose = vi.fn();
    const onFilterClose = vi.fn();
    const { rerender } = render(<KeyboardShortcutsModal open onClose={onKeyboardClose} />);

    expect(screen.getByRole('dialog', { name: 'Keyboard Shortcuts' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Close keyboard shortcuts' })).toBeInTheDocument();
    fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' });
    expect(onKeyboardClose).toHaveBeenCalledOnce();

    rerender(
      <FilterDrawer open onClose={onFilterClose}>
        <label>
          Status
          <select aria-label="Status">
            <option>Open</option>
          </select>
        </label>
      </FilterDrawer>
    );

    expect(screen.getByRole('dialog', { name: 'Filters' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Close filters' })).toBeInTheDocument();
    fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' });
    expect(onFilterClose).toHaveBeenCalledOnce();
  });
});
