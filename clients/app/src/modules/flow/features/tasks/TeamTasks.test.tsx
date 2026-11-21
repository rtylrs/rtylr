import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';

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
vi.mock('../../shared/lib/currentUser', () => ({ useCurrentUserId: () => '1' }));
vi.mock('./useLabels', () => ({ useLabels: () => ({ data: [] }) }));
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
  ...extra,
});
const listCalls = vi.hoisted(() => [] as unknown[]);
vi.mock('./useTasks', () => ({
  useTasks: (params: unknown) => {
    listCalls.push(params);
    return {
      isPending: false,
      isError: false,
      refetch: vi.fn(),
      data: [
        row('1', 'Advisor chat', {
          assigneeId: 2,
          assigneeName: 'Ahmad Tarabein',
          dueDate: '2026-12-01',
        }),
        row('2', 'Payment plan', {
          assigneeId: 2,
          assigneeName: 'Ahmad Tarabein',
          dueDate: '2026-11-01',
        }),
        row('3', 'Backup restore test', { assigneeId: 1, assigneeName: 'Abed Amouneh' }),
        row('4', 'Pick fonts'),
        row('5', 'Old release', { assigneeId: 1, assigneeName: 'Abed Amouneh', status: 'done' }),
      ],
    };
  },
}));

import TeamTasks from './TeamTasks';

const sections = (): string[] =>
  [...document.querySelectorAll<HTMLElement>('[data-team-section]')].map(
    (el) => el.dataset.teamSection ?? ''
  );
const rowsIn = (id: string): string[] =>
  [
    ...(document.querySelector(`[data-team-section="${id}"]`)?.querySelectorAll('[data-row]') ??
      []),
  ].map((el) => el.textContent ?? '');

beforeAll(async () => {
  await initTestI18n();
});

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  router.set({});
  listCalls.length = 0;
});

describe('Team tasks', () => {
  it('asks for every visible task, not one project', () => {
    render(<TeamTasks />);
    expect(listCalls[0]).toEqual({});
  });

  it('groups open tasks by person: me first, Unassigned last, soonest due first', () => {
    render(<TeamTasks />);
    expect(sections()).toEqual(['1', '2', 'none']);
    const mine = document.querySelector('[data-team-section="1"]') as HTMLElement;
    expect(within(mine).getByText('Abed Amouneh (you)')).toBeTruthy();
    expect(rowsIn('2')).toEqual(['Payment plan', 'Advisor chat']);
    expect(rowsIn('1')).toEqual(['Backup restore test']);
  });

  it("shows one person's tasks from the avatar strip", () => {
    render(<TeamTasks />);
    const strip = screen.getByRole('group', { name: /show tasks by person/i });
    fireEvent.click(within(strip).getByRole('button', { name: /ahmad tarabein/i }));
    expect(sections()).toEqual(['2']);
    expect(screen.getByText('2 of 4 tasks visible')).toBeTruthy();
  });

  it('adds finished work when Show completed is on', () => {
    render(<TeamTasks />);
    fireEvent.click(screen.getByRole('button', { name: /show completed/i }));
    expect(rowsIn('1')).toContain('Old release');
  });
});
