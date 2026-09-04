import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { TaskRow, StatusGroup } from './ProjectList';
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

  it('navigates to the task detail page when Open task is chosen', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectList.moreOptions' }));
    fireEvent.click(screen.getByRole('menuitem', { name: 'flow.taskMenu.openTask' }));

    expect(mockNavigate).toHaveBeenCalledWith({ to: '/task/t1' });
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
