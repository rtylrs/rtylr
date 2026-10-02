// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';

const { deleteMutate, canRemove } = vi.hoisted(() => ({
  deleteMutate: vi.fn(),
  canRemove: { value: true },
}));

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('@voxire/utils', () => ({ useClickOutside: () => undefined }));

vi.mock('../../features/tasks/useTasks', () => ({
  useDeleteTask: () => ({ mutate: deleteMutate, isPending: false }),
}));

vi.mock('../lib/currentUser', () => ({
  useCanDeleteResource: () => canRemove.value,
}));

// The real ConfirmDialog renders through @voxire/ui's Dialog; this keeps the
// test about whether a confirmation is required, not about the dialog's own
// markup, which its own tests cover.
vi.mock('./ConfirmDialog', () => ({
  default: ({
    open,
    onConfirm,
    description,
  }: {
    open: boolean;
    onConfirm: () => void;
    description: React.ReactNode;
  }) =>
    open ? (
      <div role="dialog">
        <div>{description}</div>
        <button type="button" onClick={onConfirm}>
          confirm-delete
        </button>
      </div>
    ) : null,
}));

import { TaskContextMenu } from './TaskContextMenu';

const menu = { x: 10, y: 20, taskId: '998504', projectId: '900', taskTitle: 'Wire up auth flow' };

const renderMenu = (onClose = vi.fn()) => render(<TaskContextMenu menu={menu} onClose={onClose} />);

describe('TaskContextMenu deletion (#1056)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    canRemove.value = true;
  });

  it('does not delete on the first click', () => {
    renderMenu();

    fireEvent.click(screen.getByRole('menuitem', { name: /flowCommon.delete/ }));

    expect(deleteMutate).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('names the task in the confirmation', () => {
    renderMenu();

    fireEvent.click(screen.getByRole('menuitem', { name: /flowCommon.delete/ }));

    expect(screen.getByText('Wire up auth flow')).toBeInTheDocument();
  });

  it('deletes once the confirmation is accepted', () => {
    renderMenu();

    fireEvent.click(screen.getByRole('menuitem', { name: /flowCommon.delete/ }));
    fireEvent.click(screen.getByRole('button', { name: 'confirm-delete' }));

    expect(deleteMutate).toHaveBeenCalledWith('998504', expect.any(Object));
  });

  it('offers no confirmation to someone who may not delete', () => {
    canRemove.value = false;
    renderMenu();

    // The name still comes from the item's text; the restriction is carried by
    // aria-disabled and the title, which do not override it.
    const deleteItem = screen.getByRole('menuitem', { name: /flowCommon.delete/ });
    expect(deleteItem).toHaveAttribute('title', 'flowCommon.deleteRestricted');

    fireEvent.click(deleteItem);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(deleteMutate).not.toHaveBeenCalled();
  });
});

describe('TaskContextMenu menu semantics (#1050)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    canRemove.value = true;
  });

  it('exposes its items as menu items', () => {
    renderMenu();

    expect(screen.getAllByRole('menuitem')).toHaveLength(3);
  });

  it('translates every label instead of hardcoding English', () => {
    renderMenu();

    const names = screen.getAllByRole('menuitem').map((item) => item.textContent);
    expect(names).toEqual([
      'flow.taskMenu.openTask',
      'flow.taskMenu.copyLink',
      'flowCommon.delete',
    ]);
  });

  it('gives the menu an accessible name', () => {
    renderMenu();

    expect(screen.getByRole('menu')).toHaveAccessibleName('flow.taskMenu.label');
  });

  it('focuses the first item and moves with the arrow keys', () => {
    renderMenu();
    const items = screen.getAllByRole('menuitem');

    expect(items[0]).toHaveFocus();

    fireEvent.keyDown(screen.getByRole('menu'), { key: 'ArrowDown' });
    expect(items[1]).toHaveFocus();

    fireEvent.keyDown(screen.getByRole('menu'), { key: 'End' });
    expect(items[2]).toHaveFocus();

    // Wraps, so the sequence is a loop rather than a dead end.
    fireEvent.keyDown(screen.getByRole('menu'), { key: 'ArrowDown' });
    expect(items[0]).toHaveFocus();

    fireEvent.keyDown(screen.getByRole('menu'), { key: 'ArrowUp' });
    expect(items[2]).toHaveFocus();
  });

  it('keeps a restricted Delete in the arrow sequence', () => {
    canRemove.value = false;
    renderMenu();
    const items = screen.getAllByRole('menuitem');

    // aria-disabled rather than disabled, so it is announced and reachable.
    expect(items[2]).toHaveAttribute('aria-disabled', 'true');
    expect(items[2]).not.toBeDisabled();

    fireEvent.keyDown(screen.getByRole('menu'), { key: 'End' });
    expect(items[2]).toHaveFocus();
  });

  it('uses a roving tabIndex rather than putting every item in the tab order', () => {
    renderMenu();
    const items = screen.getAllByRole('menuitem');

    expect(items.map((item) => item.getAttribute('tabindex'))).toEqual(['0', '-1', '-1']);
  });
});
