import { useState } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useDialogFocus } from './useDialogFocus';

interface HarnessProps {
  open: boolean;
  onClose: () => void;
}

const Harness = ({ open, onClose }: HarnessProps) => {
  const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose });

  if (!open) return null;

  return (
    <div ref={dialogRef} tabIndex={-1} data-testid="dialog">
      <button type="button">First action</button>
      <button type="button">Last action</button>
    </div>
  );
};

describe('useDialogFocus', () => {
  it('focuses the first control, traps Tab, and restores the trigger', () => {
    const trigger = document.createElement('button');
    trigger.textContent = 'Open dialog';
    document.body.append(trigger);
    trigger.focus();

    const onClose = vi.fn();
    const { rerender } = render(<Harness open onClose={onClose} />);
    const first = screen.getByRole('button', { name: 'First action' });
    const last = screen.getByRole('button', { name: 'Last action' });

    expect(first).toHaveFocus();

    last.focus();
    fireEvent.keyDown(last, { key: 'Tab' });
    expect(first).toHaveFocus();

    first.focus();
    fireEvent.keyDown(first, { key: 'Tab', shiftKey: true });
    expect(last).toHaveFocus();

    fireEvent.keyDown(last, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledOnce();

    rerender(<Harness open={false} onClose={onClose} />);
    expect(trigger).toHaveFocus();
    trigger.remove();
  });

  it('keeps focus on the dialog when it has no focusable controls', () => {
    const onClose = vi.fn();

    const EmptyHarness = () => {
      const [open] = useState(true);
      const dialogRef = useDialogFocus<HTMLDivElement>({ open, onClose });
      return <div ref={dialogRef} tabIndex={-1} data-testid="empty-dialog" />;
    };

    render(<EmptyHarness />);
    const dialog = screen.getByTestId('empty-dialog');

    expect(dialog).toHaveFocus();
    fireEvent.keyDown(dialog, { key: 'Tab' });
    expect(dialog).toHaveFocus();
  });
});
