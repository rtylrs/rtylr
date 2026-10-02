// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({
    t: (key: string, options?: Record<string, unknown>) =>
      options ? `${key}|${JSON.stringify(options)}` : key,
  }),
}));

// The real Dialog is a Radix portal with its own focus handling; rendering the
// content inline keeps these tests about this component's own markup.
vi.mock('@voxire/ui', () => ({
  Dialog: ({ open, children }: { open: boolean; children: React.ReactNode }) =>
    open ? <div>{children}</div> : null,
  DialogContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogHeader: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  DialogTitle: ({ children }: { children: React.ReactNode }) => <h2>{children}</h2>,
  DialogDescription: ({ children }: { children: React.ReactNode }) => <p>{children}</p>,
}));

import ConfirmDialog from './ConfirmDialog';

const baseProps = {
  open: true,
  onOpenChange: vi.fn(),
  title: 'Delete permanently',
  description: 'This cannot be undone.',
  confirmLabel: 'Delete permanently',
  cancelLabel: 'Cancel',
  onConfirm: vi.fn(),
};

describe('ConfirmDialog type-to-confirm (#1062)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('gives the typed input an accessible name', () => {
    render(<ConfirmDialog {...baseProps} requireText="delete" />);

    // Before this the field had no label, no aria-label, no aria-describedby
    // and no id: an unnamed text box next to a disabled destructive button.
    expect(
      screen.getByLabelText('flow.confirm.typeToConfirm|{"word":"delete"}')
    ).toBeInTheDocument();
  });

  it('states the requirement rather than leaving it to the placeholder', () => {
    render(<ConfirmDialog {...baseProps} requireText="delete" />);

    expect(screen.getByText('flow.confirm.typeToConfirm|{"word":"delete"}')).toBeInTheDocument();
  });

  it('keeps confirm disabled until the exact word is typed', () => {
    render(<ConfirmDialog {...baseProps} requireText="delete" />);
    const confirm = screen.getByRole('button', { name: 'Delete permanently' });
    const input = screen.getByLabelText('flow.confirm.typeToConfirm|{"word":"delete"}');

    expect(confirm).toBeDisabled();

    fireEvent.change(input, { target: { value: 'del' } });
    expect(confirm).toBeDisabled();

    fireEvent.change(input, { target: { value: '  delete  ' } });
    expect(confirm).not.toBeDisabled();
  });

  it('renders no input when no word is required', () => {
    render(<ConfirmDialog {...baseProps} />);

    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Delete permanently' })).not.toBeDisabled();
  });

  it('translates its pending label', () => {
    render(<ConfirmDialog {...baseProps} pending />);

    // This read the literal English 'Working…' in a dialog used by four
    // screens across three languages.
    expect(screen.getByRole('button', { name: 'flow.confirm.working' })).toBeInTheDocument();
  });
});
