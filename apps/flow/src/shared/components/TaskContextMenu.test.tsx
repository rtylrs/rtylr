import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// @voxire/utils is deliberately NOT mocked: useClickOutside is the thing under
// test here. Everything else this component pulls in is stubbed.
const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  deleteMutate: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

vi.mock('../../features/tasks/useTasks', () => ({
  useDeleteTask: () => ({ mutate: mocks.deleteMutate }),
}));

vi.mock('../lib/currentUser', () => ({
  useCanDeleteResource: () => true,
}));

import { TaskContextMenu } from './TaskContextMenu';

const MENU = { x: 10, y: 20, taskId: 't1', projectId: 'p1' };

function renderMenu() {
  const onClose = vi.fn();
  render(<TaskContextMenu menu={MENU} onClose={onClose} />);
  return { onClose };
}

describe('TaskContextMenu dismissal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('closes on an outside mousedown', () => {
    const { onClose } = renderMenu();
    fireEvent.mouseDown(document.body);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('stays open on a mousedown inside the menu', () => {
    const { onClose } = renderMenu();
    fireEvent.mouseDown(screen.getByRole('menu'));
    expect(onClose).not.toHaveBeenCalled();
  });

  // The hand-rolled listener this replaced bound 'mousedown' only, so tapping
  // outside the menu on a touch device left it open. The shared hook binds
  // 'touchstart' too.
  it('closes on an outside touchstart', () => {
    const { onClose } = renderMenu();
    fireEvent.touchStart(document.body);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('closes on Escape', () => {
    const { onClose } = renderMenu();
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});
