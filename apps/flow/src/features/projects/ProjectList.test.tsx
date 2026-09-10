import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { TaskRow, StatusGroup, STATUS_GROUP_COLORS, groupIsVisible } from './ProjectList';
import StatusIcon from '../../shared/components/StatusIcon';
import { STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import type { Task } from '../../shared/stores/flowStore';

const mockNavigate = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({}),
  Link: ({ children }: { children?: ReactNode }) => <>{children}</>,
}));

// Identity translator: renders the i18n key itself instead of resolved text,
// same convention used in TaskDetail.test.tsx. What matters for the
// regression test below is that both components resolve the SAME key for
// the same status, not the literal English string. Keep the rest of the
// module real (via importOriginal) - `initReactI18next` is still needed by
// the real i18next instance that useTasks.ts imports as a side effect.
vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({ t: (key: string) => key }),
  };
});

const task: Task = {
  id: 't1',
  projectId: 'p-web',
  workspaceId: 'w1',
  title: 'Ship the release',
  status: 'todo',
  priority: 'medium',
  assignee: 'Ada Lovelace',
  assigneeInitials: 'AL',
  assigneeId: 'u1',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Web',
  createdAt: '',
  updatedAt: '',
  identifier: 'FLOW-1',
};

function renderRow(): void {
  const queryClient = new QueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <TaskRow task={task} />
    </QueryClientProvider>
  );
}

describe('ProjectList TaskRow more menu', () => {
  // The labels are translation keys now: they were literal English until the
  // menu was internationalised, and this suite's `t` mock returns the key.
  it('opens a menu with open/copy link/delete actions when the more button is clicked', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectList.moreOptions' }));

    expect(screen.getByRole('menuitem', { name: 'flow.taskMenu.openTask' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: 'flow.taskMenu.copyLink' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: 'common.delete' })).toBeInTheDocument();
  });

  it('opens the task over the list when Open task is chosen', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectList.moreOptions' }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'flow.taskMenu.openTask' }));

    // `?openTask`, not `/task/t1`: the list stays mounted and the drawer opens
    // over it (rtylr#1079).
    const call = mockNavigate.mock.calls[0][0] as {
      to: string;
      search: (previous: Record<string, unknown>) => Record<string, unknown>;
    };
    expect(call.to).toBe('.');
    expect(call.search({})).toEqual({ openTask: 't1' });
  });
});

describe('Status label consistency (issue #756)', () => {
  // The project List view (StatusGroup) and any other view that shows a task
  // status (StatusIcon, the task detail status picker) must render the exact
  // same text for the same status. Previously the List view hardcoded its
  // own "Todo" while the rest of the app used "To Do" via i18n. Both now
  // resolve through the single STATUS_LABEL_KEYS map, so this asserts they
  // stay identical rather than re-hardcoding a string that could drift again.
  it('renders the same "todo" label in the List view group header and StatusIcon', () => {
    const todoLabel = STATUS_LABEL_KEYS.todo;

    const { container: groupContainer } = render(
      <StatusGroup group={{ id: 'todo', label: todoLabel, color: '#6b7280' }} tasks={[]} />
    );
    const { container: iconContainer } = render(<StatusIcon status="todo" showLabel />);

    expect(groupContainer.textContent).toContain(todoLabel);
    expect(iconContainer.textContent).toContain(todoLabel);
  });

  it('has a distinct label key for every status shown in the List view', () => {
    const listStatuses: Task['status'][] = ['todo', 'in_progress', 'in_review', 'done'];
    const keys = listStatuses.map((status) => STATUS_LABEL_KEYS[status]);

    expect(new Set(keys).size).toBe(listStatuses.length);
  });
});

describe('ProjectList status groups (#1049: a backlog task had no home)', () => {
  const byId = (id: Task['status']) => STATUS_GROUP_COLORS.find((g) => g.id === id);

  it('has a group for every value of the status enum', () => {
    // The list offered four groups against a six-value enum, so a task in
    // backlog or cancelled matched no group and rendered nowhere, while the
    // board showed it in its first column.
    expect(STATUS_GROUP_COLORS.map((g) => g.id)).toEqual([
      'backlog',
      'todo',
      'in_progress',
      'in_review',
      'done',
      'cancelled',
    ]);
  });

  it('covers every status the label map knows about', () => {
    // STATUS_LABEL_KEYS is the app's single source of status vocabulary, so
    // any status it names has to be groupable here.
    expect(STATUS_GROUP_COLORS.map((g) => g.id).sort()).toEqual(
      Object.keys(STATUS_LABEL_KEYS).sort()
    );
  });

  it('shows a backlog or cancelled group only once it holds something', () => {
    expect(groupIsVisible(byId('backlog')!, 0)).toBe(false);
    expect(groupIsVisible(byId('backlog')!, 1)).toBe(true);
    expect(groupIsVisible(byId('cancelled')!, 0)).toBe(false);
    expect(groupIsVisible(byId('cancelled')!, 3)).toBe(true);
  });

  it('keeps the four familiar groups on screen when they are empty', () => {
    // A project that has neither backlog nor cancelled work must look exactly
    // as it did before, four headers and no more.
    for (const id of ['todo', 'in_progress', 'in_review', 'done'] as Task['status'][]) {
      expect(groupIsVisible(byId(id)!, 0)).toBe(true);
    }
  });

  it('puts backlog before todo and cancelled last', () => {
    const order = STATUS_GROUP_COLORS.map((g) => g.id);
    expect(order.indexOf('backlog')).toBeLessThan(order.indexOf('todo'));
    expect(order.indexOf('cancelled')).toBe(order.length - 1);
  });
});

describe('StatusGroup count scope (#1166)', () => {
  const tasks = [
    { id: 't1', title: 'One', status: 'todo' },
    { id: 't2', title: 'Two', status: 'todo' },
  ] as unknown as Parameters<typeof StatusGroup>[0]['tasks'];

  const group = { id: 'todo', color: '#9ca3af', label: 'To Do' } as unknown as Parameters<
    typeof StatusGroup
  >[0]['group'];

  const renderGroup = (showCount?: boolean): void => {
    const queryClient = new QueryClient();
    render(
      <QueryClientProvider client={queryClient}>
        <StatusGroup group={group} tasks={tasks} showCount={showCount} />
      </QueryClientProvider>
    );
  };

  it('shows the count when the list is not split into sections', () => {
    // Here the group count and the board's count are the same number, so there
    // is nothing to confuse and the number is worth keeping.
    renderGroup();
    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('hides the count when the list is sectioned', () => {
    // The section total one line above already gives the number, and this one
    // counts only the section while the board's identical label counts the
    // whole project.
    renderGroup(false);
    expect(screen.queryByText('2')).not.toBeInTheDocument();
  });

  it('still renders the status label either way', () => {
    renderGroup(false);
    expect(screen.getByText('To Do')).toBeInTheDocument();
  });
});

describe('the sectioned list actually asks for the count to be hidden', () => {
  it('passes showCount={false} only in the sectioned branch', () => {
    // Testing StatusGroup alone proves the prop works, not that anything sets
    // it. Deleting the prop at the call site left every component test green.
    const source = fs.readFileSync(path.resolve(__dirname, 'ProjectList.tsx'), 'utf8');

    const sectioned = source.slice(
      source.indexOf('block.tasks.filter'),
      source.indexOf('projectTasks.filter((t) => t.status === group.id)')
    );
    expect(sectioned, 'sliced the wrong region').toContain('StatusGroup');
    expect(sectioned, 'the sectioned list must hide the group count').toContain(
      'showCount={false}'
    );

    const unsectioned = source.slice(
      source.indexOf('projectTasks.filter((t) => t.status === group.id)')
    );
    expect(unsectioned, 'sliced the wrong region').toContain('StatusGroup');
    expect(unsectioned, 'an unsectioned list keeps its count').not.toContain('showCount={false}');
  });
});
