// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import MoveToColumnMenu from './MoveToColumnMenu';

const targets = [
  { id: '1', name: 'Backlog' },
  { id: '2', name: 'In progress' },
  { id: '3', name: 'Done' },
];

function setup(overrides: Partial<React.ComponentProps<typeof MoveToColumnMenu>> = {}) {
  const onMove = vi.fn();
  render(
    <MoveToColumnMenu
      taskTitle="Ship it"
      targets={targets}
      currentColumnId="1"
      onMove={onMove}
      {...overrides}
    />
  );
  return { onMove };
}

const trigger = (): HTMLElement => screen.getByRole('button', { name: 'flow.board.moveTask' });

describe('MoveToColumnMenu', () => {
  beforeEach(() => vi.clearAllMocks());

  // WCAG 2.2 AA 2.5.7: the drag has to have a non-drag equivalent, and it has
  // to be announced as a menu rather than an anonymous icon.
  it('exposes the trigger as a collapsed menu button with a name', () => {
    setup();
    expect(trigger()).toHaveAttribute('aria-haspopup', 'menu');
    expect(trigger()).toHaveAttribute('aria-expanded', 'false');
  });

  it('reports expanded state once open', () => {
    setup();
    fireEvent.click(trigger());
    expect(trigger()).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByRole('menu')).toBeInTheDocument();
  });

  it('moves the task to the chosen column and closes', () => {
    const { onMove } = setup();
    fireEvent.click(trigger());
    fireEvent.click(screen.getByRole('menuitem', { name: /In progress/ }));
    expect(onMove).toHaveBeenCalledWith('2');
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  // Dropping a card on the column it already sits in is not a move, so the
  // menu must not offer it as one.
  it('marks the current column and does not let it be chosen', () => {
    const { onMove } = setup();
    fireEvent.click(trigger());
    const current = screen.getByRole('menuitem', { name: /Backlog/ });
    expect(current).toBeDisabled();
    expect(current).toHaveAttribute('aria-current', 'true');
    fireEvent.click(current);
    expect(onMove).not.toHaveBeenCalled();
  });

  // The current column renders disabled and is usually first, so "focus the
  // first item" has to mean the first item that can actually take focus.
  it('focuses the first selectable item on open, skipping the current column', () => {
    setup();
    fireEvent.click(trigger());
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: /In progress/ }));
  });

  it('walks the selectable items with arrow keys and wraps around', () => {
    setup();
    fireEvent.click(trigger());
    const menu = screen.getByRole('menu');
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: /In progress/ }));
    fireEvent.keyDown(menu, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: /Done/ }));
    fireEvent.keyDown(menu, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: /In progress/ }));
    fireEvent.keyDown(menu, { key: 'ArrowUp' });
    expect(document.activeElement).toBe(screen.getByRole('menuitem', { name: /Done/ }));
  });

  it('closes on Escape', () => {
    setup();
    fireEvent.click(trigger());
    fireEvent.keyDown(screen.getByRole('menu'), { key: 'Escape' });
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('closes on an outside press', () => {
    setup();
    fireEvent.click(trigger());
    fireEvent.pointerDown(document.body);
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });

  it('says so rather than opening an empty menu when there is nowhere to move', () => {
    setup({ targets: [], currentColumnId: null });
    fireEvent.click(trigger());
    expect(screen.getByText('flow.board.moveNoTargets')).toBeInTheDocument();
    expect(screen.queryAllByRole('menuitem')).toHaveLength(0);
  });

  it('cannot be opened without edit rights', () => {
    setup({ disabled: true });
    expect(trigger()).toBeDisabled();
    fireEvent.click(trigger());
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
  });
});
