import React, { useState } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useDialogBehavior } from './useDialogBehavior';

/**
 * These drive the behaviour rather than asserting on the source, because the
 * bug was precisely that a handler existed and could never run: `onKeyDown`
 * sat on an overlay div with no tabIndex, and focus was never moved into the
 * dialog, so the keydown landed on document.body and never bubbled through.
 */
function Dialog({
  open,
  onClose,
  empty = false,
}: {
  open: boolean;
  onClose: () => void;
  empty?: boolean;
}): React.JSX.Element | null {
  const { panelRef } = useDialogBehavior<HTMLDivElement>(open, onClose);
  if (!open) return null;
  return (
    <div role="dialog" aria-modal="true" aria-label="Settings">
      <div ref={panelRef} tabIndex={-1} data-testid="panel">
        {!empty && (
          <>
            <button type="button">First</button>
            <input aria-label="Middle" />
            <button type="button">Last</button>
          </>
        )}
      </div>
    </div>
  );
}

describe('useDialogBehavior', () => {
  it('moves focus into the panel on open', async () => {
    render(<Dialog open onClose={vi.fn()} />);
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'First' }))
    );
  });

  it('focuses the panel itself when it holds nothing focusable', async () => {
    render(<Dialog open empty onClose={vi.fn()} />);
    await waitFor(() => expect(document.activeElement).toBe(screen.getByTestId('panel')));
  });

  it('closes on Escape with focus inside the dialog', async () => {
    const onClose = vi.fn();
    render(<Dialog open onClose={onClose} />);
    await waitFor(() => expect(document.activeElement).not.toBe(document.body));
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('closes on Escape when the key lands on the body', () => {
    // The reported case. The old handler was on an unfocusable overlay div, so
    // a keydown outside it never reached the dialog and Esc did nothing.
    const onClose = vi.fn();
    render(<Dialog open onClose={onClose} />);
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('ignores Escape while closed', () => {
    const onClose = vi.fn();
    render(<Dialog open={false} onClose={onClose} />);
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).not.toHaveBeenCalled();
  });

  it('cycles Tab from the last control back to the first', async () => {
    render(<Dialog open onClose={vi.fn()} />);
    const last = screen.getByRole('button', { name: 'Last' });
    last.focus();
    fireEvent.keyDown(last, { key: 'Tab' });
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'First' }))
    );
  });

  it('cycles Shift+Tab from the first control back to the last', async () => {
    render(<Dialog open onClose={vi.fn()} />);
    const first = screen.getByRole('button', { name: 'First' });
    first.focus();
    fireEvent.keyDown(first, { key: 'Tab', shiftKey: true });
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Last' }))
    );
  });

  it('leaves Tab alone in the middle of the panel', () => {
    render(<Dialog open onClose={vi.fn()} />);
    const middle = screen.getByLabelText('Middle');
    middle.focus();
    // Not prevented, so the browser's own order applies inside the panel.
    const prevented = !fireEvent.keyDown(middle, { key: 'Tab' });
    expect(prevented).toBe(false);
    expect(document.activeElement).toBe(middle);
  });

  it('restores focus to the opener on close', async () => {
    function Host(): React.JSX.Element {
      const [open, setOpen] = useState(false);
      return (
        <>
          <button type="button" onClick={() => setOpen(true)}>
            Open
          </button>
          <Dialog open={open} onClose={() => setOpen(false)} />
        </>
      );
    }
    render(<Host />);
    const opener = screen.getByRole('button', { name: 'Open' });
    opener.focus();
    fireEvent.click(opener);
    await waitFor(() =>
      expect(document.activeElement).toBe(screen.getByRole('button', { name: 'First' }))
    );
    fireEvent.keyDown(document.activeElement as HTMLElement, { key: 'Escape' });
    await waitFor(() => expect(document.activeElement).toBe(opener));
  });
});
