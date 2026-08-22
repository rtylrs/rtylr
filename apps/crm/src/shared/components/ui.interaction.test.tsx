import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';
import { ConfirmModal, Modal } from './ui';

describe('Modal focus behavior', () => {
  it('keeps focus in a controlled field when the parent rerenders', async () => {
    const user = userEvent.setup();

    const Harness: React.FC = () => {
      const [value, setValue] = useState('');
      return (
        <Modal open onClose={() => undefined} title="Edit customer">
          <input
            aria-label="Customer name"
            value={value}
            onChange={(event) => setValue(event.target.value)}
          />
        </Modal>
      );
    };

    render(<Harness />);
    const input = screen.getByRole('textbox', { name: 'Customer name' });
    await user.click(input);
    fireEvent.change(input, { target: { value: 'Ada' } });

    expect(input).toHaveFocus();
  });

  it('restores focus to the opener when it closes', async () => {
    const user = userEvent.setup();

    const Harness: React.FC = () => {
      const [open, setOpen] = useState(false);
      return (
        <>
          <button type="button" onClick={() => setOpen(true)}>
            Open modal
          </button>
          <Modal open={open} onClose={() => setOpen(false)} title="Confirm">
            Ready
          </Modal>
        </>
      );
    };

    render(<Harness />);
    const opener = screen.getByRole('button', { name: 'Open modal' });
    await user.click(opener);
    await user.click(screen.getByRole('button', { name: 'Close' }));

    expect(opener).toHaveFocus();
  });

  it('keeps keyboard focus inside the dialog', async () => {
    const user = userEvent.setup();

    render(
      <Modal
        open
        onClose={() => undefined}
        title="Save customer"
        footer={
          <button type="button" onClick={() => undefined}>
            Save
          </button>
        }
      >
        <input aria-label="Customer name" />
      </Modal>
    );

    const close = screen.getByRole('button', { name: 'Close' });
    const input = screen.getByRole('textbox', { name: 'Customer name' });
    const save = screen.getByRole('button', { name: 'Save' });

    expect(close).toHaveFocus();
    await user.tab();
    expect(input).toHaveFocus();
    await user.tab();
    expect(save).toHaveFocus();
    await user.tab();
    expect(close).toHaveFocus();
    await user.tab({ shift: true });
    expect(save).toHaveFocus();
  });

  it('can keep an async confirmation open until the mutation succeeds', async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    const Harness: React.FC = () => {
      const [open, setOpen] = useState(true);
      return (
        <ConfirmModal
          open={open}
          onClose={() => setOpen(false)}
          onConfirm={onConfirm}
          closeOnConfirm={false}
          title="Delete customer"
        />
      );
    };

    render(<Harness />);
    await user.click(screen.getByRole('button', { name: 'Confirm' }));

    expect(onConfirm).toHaveBeenCalledOnce();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });
});
