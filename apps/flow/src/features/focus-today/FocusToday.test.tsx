import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children?: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => navigateMock,
}));

const myIssuesState = vi.hoisted(() => ({ data: [] as ApiTask[] }));
vi.mock('../tasks/useMyIssues', () => ({
  useMyIssues: () => myIssuesState,
}));

// activeWorkspaceId is '' in production: the workspace layer is gone. The old
// filter compared it against each task's workspaceId, which the API no longer
// sends, so a mock carrying 'ws-1' hid the defect in #1035.
// importOriginal, because rendering a task pulls PRIORITY_COLORS and friends
// out of this module; only the hook is replaced.
vi.mock('../../shared/stores/flowStore', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/stores/flowStore')>()),
  useFlowStore: () => ({ projects: [], activeWorkspaceId: '' }),
}));

// The existing tests below all render with no tasks, so this only affects the
// bucketing cases added at the end of the file.
vi.mock('../../shared/lib/currentUser', () => ({ isTaskMine: () => true }));

import FocusToday from './FocusToday';

beforeEach(() => {
  vi.clearAllMocks();
  myIssuesState.data = [];
});

describe('FocusToday empty state', () => {
  it('shows the empty title/hint with a CTA to My Tasks', () => {
    render(<FocusToday />);

    expect(screen.getByText('common.focusToday.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('common.focusToday.emptyHint')).toBeInTheDocument();

    const cta = screen.getByRole('button', { name: 'flow.shell.myTasks' });
    fireEvent.click(cta);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/my-issues' });
  });
});

describe('FocusToday lists what is due today (#1035)', () => {
  const isoToday = (): string => {
    const d = new Date();
    d.setHours(12, 0, 0, 0);
    return d.toISOString();
  };

  const task = (over: Partial<ApiTask> & { id: string; title: string }): ApiTask =>
    ({ status: 'todo', assigneeId: '9', dueDate: '', ...over }) as unknown as ApiTask;

  it('lists a task due today', () => {
    myIssuesState.data = [task({ id: '1', title: 'Due today thing', dueDate: isoToday() })];

    render(<FocusToday />);

    expect(screen.getByText('Due today thing')).toBeInTheDocument();
  });

  it('lists a pinned task with no due date', () => {
    myIssuesState.data = [task({ id: '2', title: 'Pinned thing', focusPin: true })];

    render(<FocusToday />);

    expect(screen.getByText('Pinned thing')).toBeInTheDocument();
  });

  it('leaves out a task that is neither pinned nor due today', () => {
    myIssuesState.data = [task({ id: '3', title: 'Some other week', dueDate: '2027-01-01' })];

    render(<FocusToday />);

    expect(screen.queryByText('Some other week')).not.toBeInTheDocument();
  });

  it('leaves out finished and archived work due today', () => {
    myIssuesState.data = [
      task({ id: '4', title: 'Done today', status: 'done', dueDate: isoToday() }),
      task({ id: '5', title: 'Cancelled today', status: 'cancelled', dueDate: isoToday() }),
      task({ id: '6', title: 'Archived today', dueDate: isoToday(), archived: true }),
      task({ id: '7', title: 'Open today', dueDate: isoToday() }),
    ];

    render(<FocusToday />);

    expect(screen.getByText('Open today')).toBeInTheDocument();
    expect(screen.queryByText('Done today')).not.toBeInTheDocument();
    expect(screen.queryByText('Cancelled today')).not.toBeInTheDocument();
    expect(screen.queryByText('Archived today')).not.toBeInTheDocument();
  });

  it('puts pinned work above work merely due today', () => {
    myIssuesState.data = [
      task({ id: '8', title: 'Just due', dueDate: isoToday() }),
      task({ id: '9', title: 'Actually pinned', focusPin: true }),
    ];

    const { container } = render(<FocusToday />);

    const text = container.textContent ?? '';
    expect(text.indexOf('Actually pinned')).toBeLessThan(text.indexOf('Just due'));
  });
});
