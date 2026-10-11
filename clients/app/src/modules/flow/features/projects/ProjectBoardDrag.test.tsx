import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectBoard from './ProjectBoard';
import { initTestI18n } from '../../shared/testing/i18n';

const access = vi.hoisted(() => ({ open: vi.fn() }));
const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: '2' }),
  useLocation: () => ({ pathname: '/flow/p/2' }),
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanCreateTask: () => true,
  useCanEditTask: () => true,
  useCanDeleteResource: () => true,
  useCanAdministerProject: () => true,
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => access.open }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '2', name: 'Release project', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
  }),
}));

const task = (id: string, title: string, sectionId: string, sortOrder?: number) => ({
  id,
  projectId: '2',
  sectionId,
  sortOrder,
  title,
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Release project',
});
const taskData = vi.hoisted(() => ({ rows: [] as unknown[] }));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({ isLoading: false, isError: false, data: taskData.rows }),
}));
vi.mock('../tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('./useColumns', () => ({
  useProjectColumns: () => ({
    isLoading: false,
    isError: false,
    data: [
      { id: '9', name: 'To do', color: '#6366f1', sortOrder: 0 },
      { id: '10', name: 'Doing', color: '#f59e0b', sortOrder: 1 },
    ],
  }),
  useCreateColumn: () => mutation,
  useUpdateColumn: () => mutation,
  useDeleteColumn: () => mutation,
  useMoveTaskToColumn: () => mutation,
}));
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({ data: [], isPending: false, isError: false }),
  useCreateStatus: () => mutation,
  useUpdateStatus: () => mutation,
  useDeleteStatus: () => mutation,
}));

// happy-dom has no layout, so the geometry the drag reads is supplied here:
// two 300px columns side by side, each list starting at y=100, cards 60px tall
// with a 6px gap (centres at 130, 196, 262).
const rect = (left: number, top: number, width: number, height: number): DOMRect =>
  ({
    left,
    top,
    width,
    height,
    right: left + width,
    bottom: top + height,
    x: left,
    y: top,
    toJSON: () => ({}),
  }) as DOMRect;
const COLUMN_X: Record<string, number> = { '9': 0, '10': 320 };

function stubGeometry(): () => void {
  const original = HTMLElement.prototype.getBoundingClientRect;
  HTMLElement.prototype.getBoundingClientRect = function (this: HTMLElement): DOMRect {
    const column = this.closest<HTMLElement>('[data-board-column]');
    const x = COLUMN_X[column?.dataset.boardColumn ?? ''] ?? 0;
    if (this.dataset.boardColumn) return rect(x, 0, 300, 800);
    if (this.hasAttribute('data-board-list')) return rect(x, 100, 300, 600);
    if (this.dataset.boardCard) {
      const siblings = [...(this.parentElement?.querySelectorAll('[data-board-card]') ?? [])];
      const visible = siblings.filter((el) => (el as HTMLElement).style.position !== 'absolute');
      const at = Math.max(0, visible.indexOf(this));
      return rect(x, 100 + at * 66, 300, 60);
    }
    return rect(0, 0, 0, 0);
  };
  return () => {
    HTMLElement.prototype.getBoundingClientRect = original;
  };
}

const down = (el: Element, x: number, y: number): void => {
  fireEvent.pointerDown(el, {
    button: 0,
    clientX: x,
    clientY: y,
    pointerId: 1,
    pointerType: 'mouse',
  });
};
const move = (x: number, y: number): void => {
  fireEvent.pointerMove(window, { clientX: x, clientY: y, pointerId: 1, pointerType: 'mouse' });
};
const up = (x: number, y: number): void => {
  fireEvent.pointerUp(window, { clientX: x, clientY: y, pointerId: 1, pointerType: 'mouse' });
};
const card = (title: string): HTMLElement =>
  screen.getByRole('button', { name: new RegExp(`Open task.*${title}`) });
const placeholderIn = (columnId: string): Element | null =>
  document.querySelector(`[data-board-column="${columnId}"] [data-board-placeholder]`);
const orderIn = (columnId: string): string[] =>
  [
    ...document.querySelectorAll(
      `[data-board-column="${columnId}"] [data-board-card], [data-board-column="${columnId}"] [data-board-placeholder]`
    ),
  ]
    .filter((el) => (el as HTMLElement).style.position !== 'absolute')
    .map((el) => (el as HTMLElement).dataset.boardCard ?? 'PLACEHOLDER');

let restoreGeometry: () => void;
beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  restoreGeometry = stubGeometry();
  taskData.rows = [
    task('1', 'Task one', '9'),
    task('2', 'Task two', '9'),
    task('3', 'Task three', '9'),
    task('4', 'Task four', '10'),
  ];
});
afterEach(() => {
  restoreGeometry();
});

const renderBoard = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectBoard />
    </QueryClientProvider>
  );

describe('Board live card drag', () => {
  it('shows a placeholder in the column under the pointer and moves the others aside', () => {
    renderBoard();
    down(card('Task one'), 50, 130);
    // Not yet a drag: under the movement threshold.
    move(52, 131);
    expect(placeholderIn('9')).toBeNull();

    // Over the other column, above its only card.
    move(400, 120);
    expect(orderIn('10')).toEqual(['PLACEHOLDER', '4']);
    expect(orderIn('9')).toEqual(['PLACEHOLDER'.replace('PLACEHOLDER', '2'), '3']);

    // Further down, below it.
    move(400, 200);
    expect(orderIn('10')).toEqual(['4', 'PLACEHOLDER']);
    up(400, 200);
  });

  it('shows a floating copy of the card while dragging and removes it on release', () => {
    renderBoard();
    expect(screen.getAllByText('Task one')).toHaveLength(1);
    down(card('Task one'), 50, 130);
    move(400, 200);
    // The hidden original plus the copy that follows the pointer.
    expect(screen.getAllByText('Task one')).toHaveLength(2);
    up(400, 200);
    expect(screen.getAllByText('Task one')).toHaveLength(1);
  });

  it('saves the new place when a card is dropped in another column', () => {
    renderBoard();
    down(card('Task one'), 50, 130);
    move(400, 200);
    up(400, 200);
    expect(mutation.mutate).toHaveBeenCalledWith({
      taskId: '1',
      columnId: '10',
      orderedTaskIds: ['4', '1'],
    });
  });

  it('reorders a card inside its own column', () => {
    renderBoard();
    down(card('Task one'), 50, 130);
    // Below the second card's centre (196) but above the third's (262).
    move(60, 240);
    expect(orderIn('9')).toEqual(['2', 'PLACEHOLDER', '3']);
    up(60, 240);
    expect(mutation.mutate).toHaveBeenCalledWith({
      taskId: '1',
      columnId: '9',
      orderedTaskIds: ['2', '1', '3'],
    });
  });

  it('saves nothing when the card is dropped where it started', () => {
    renderBoard();
    down(card('Task two'), 50, 196);
    move(60, 210);
    move(60, 200);
    up(60, 200);
    expect(mutation.mutate).not.toHaveBeenCalled();
  });

  it('cancels on Escape and when released outside every column', () => {
    renderBoard();
    down(card('Task one'), 50, 130);
    move(400, 200);
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(placeholderIn('10')).toBeNull();
    up(400, 200);
    expect(mutation.mutate).not.toHaveBeenCalled();

    down(card('Task one'), 50, 130);
    move(900, 200);
    up(900, 200);
    expect(mutation.mutate).not.toHaveBeenCalled();
  });

  it('still opens a task on a plain click and not on the click that ends a drag', () => {
    renderBoard();
    down(card('Task two'), 50, 196);
    up(50, 196);
    fireEvent.click(card('Task two'));
    expect(access.open).toHaveBeenCalledWith('2');

    // A touch drag keeps the pointer on the card it started on, so its release
    // and the click that follows both land there.
    access.open.mockClear();
    // Held from before the drag: while it is dragged the original is hidden, so
    // it is no longer in the accessibility tree.
    const dragged = card('Task two');
    down(dragged, 50, 196);
    move(60, 240);
    fireEvent.pointerUp(dragged, { clientX: 60, clientY: 240, pointerId: 1 });
    fireEvent.click(dragged);
    expect(access.open).not.toHaveBeenCalled();
  });

  it('keeps cards in their saved order', () => {
    taskData.rows = [
      task('1', 'Task one', '9', 3),
      task('2', 'Task two', '9', 1),
      task('3', 'Task three', '9', 2),
      task('4', 'Task four', '10'),
    ];
    renderBoard();
    expect(orderIn('9')).toEqual(['2', '3', '1']);
  });
});
