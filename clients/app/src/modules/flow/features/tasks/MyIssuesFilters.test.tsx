import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';
import { useUserStore } from '../../shared/stores/userStore';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const router = vi.hoisted(() => {
  let search: Record<string, unknown> = {};
  const listeners = new Set<() => void>();
  return {
    get: () => search,
    set: (next: Record<string, unknown>) => {
      search = next;
      listeners.forEach((l) => l());
    },
    subscribe: (l: () => void) => {
      listeners.add(l);
      return () => listeners.delete(l);
    },
  };
});
vi.mock('@tanstack/react-router', async () => {
  const React = await import('react');
  return {
    useSearch: () => React.useSyncExternalStore(router.subscribe, router.get),
    useNavigate:
      () => (opts: { search: (prev: Record<string, unknown>) => Record<string, unknown> }) =>
        router.set(opts.search(router.get())),
  };
});
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('./useLabels', () => ({
  useLabels: () => ({ data: [{ id: 3, name: 'Design', color: '#f00', createdAt: '' }] }),
}));
vi.mock('../../shared/components/TaskRow', () => ({
  TaskRow: ({ task }: { task: { title: string } }) => <div data-row>{task.title}</div>,
}));

const row = (id: string, title: string, extra: Record<string, unknown> = {}) => ({
  id,
  projectId: '1',
  projectName: 'Investment Hub',
  title,
  status: 'todo',
  priority: 'medium',
  identifier: `T-${id}`,
  // Dated, because the no-date bucket opens collapsed.
  dueDate: '2020-01-01',
  ...extra,
});
vi.mock('./useMyIssues', () => ({
  useMyIssues: () => ({
    isPending: false,
    isError: false,
    isPaused: false,
    failureCount: 0,
    refetch: vi.fn(),
    data: [
      row('1', 'Payment plan tracker', { priority: 'high' }),
      row('2', 'Website copy', { projectId: '2', projectName: 'RTYLR', labels: ['3'] }),
      row('3', 'Shipped release notes', { status: 'done' }),
    ],
  }),
}));

import MyIssues from './MyIssues';

const rows = (): string[] =>
  [...document.querySelectorAll('[data-row]')].map((el) => el.textContent ?? '');

beforeAll(async () => {
  await initTestI18n();
});

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  router.set({});
  useUserStore.setState({
    user: { id: 7, effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
});

describe('My tasks filters', () => {
  it('searches my open tasks and counts what is left', () => {
    render(<MyIssues />);
    expect(rows().sort()).toEqual(['Payment plan tracker', 'Website copy']);
    fireEvent.change(screen.getByRole('textbox', { name: /search tasks by title/i }), {
      target: { value: 'design' },
    });
    expect(rows()).toEqual(['Website copy']);
    expect(screen.getByText('1 of 2 tasks visible')).toBeTruthy();
  });

  it('narrows to one project from a link', () => {
    router.set({ filter: { projects: ['2'] } });
    render(<MyIssues />);
    expect(rows()).toEqual(['Website copy']);
    expect(screen.getByRole('button', { name: /project: rtylr/i })).toBeTruthy();
  });

  it('shows completed work in its own section when asked, and remembers it', () => {
    render(<MyIssues />);
    expect(rows()).not.toContain('Shipped release notes');
    fireEvent.click(screen.getByRole('button', { name: /show completed/i }));
    expect(rows()).toContain('Shipped release notes');
    expect(document.querySelector('[data-my-tasks-completed]')).not.toBeNull();
    expect(localStorage.getItem('flow:my-tasks:show-completed')).toBe('1');
  });

  it('includes done tasks when the status filter asks for them', () => {
    router.set({ filter: { statuses: ['done'] } });
    render(<MyIssues />);
    expect(rows()).toEqual(['Shipped release notes']);
  });

  it('says nothing matches instead of claiming there are no tasks', () => {
    router.set({ filter: { query: 'nothing like this' } });
    render(<MyIssues />);
    expect(screen.getByText('No tasks match these filters.')).toBeTruthy();
    expect(screen.queryByText(/no tasks assigned/i)).toBeNull();
  });
});
