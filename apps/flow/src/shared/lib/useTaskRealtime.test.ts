// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { RealtimeEvent } from '@voxire/api';

// See useProjectRealtime.test.ts - same harness: capture the registered
// handler so events can be driven in without a WebSocket.
let handler: ((e: RealtimeEvent) => void) | undefined;
vi.mock('./realtime', () => ({
  useRealtime: (_topics: string[], h: (e: RealtimeEvent) => void) => {
    handler = h;
  },
}));

vi.mock('./currentUser', () => ({
  getCurrentUserId: vi.fn(() => '42'),
}));

import { useTaskRealtime } from './useTaskRealtime';
import { getCurrentUserId } from './currentUser';

function mount(taskId = 't-1') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const spy = vi.spyOn(client, 'invalidateQueries');
  renderHook(() => useTaskRealtime(taskId), {
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children),
  });
  return spy;
}

type InvalidateSpy = ReturnType<typeof mount>;

function keys(spy: InvalidateSpy): string[] {
  return spy.mock.calls.map((call) => JSON.stringify(call[0]?.queryKey));
}

const MY_TASKS = JSON.stringify(['my-tasks']);

/** A teammate-authored event (actor 999 is not the current user, 42). */
function event(overrides: Partial<RealtimeEvent> = {}): RealtimeEvent {
  return { topic: 'task:t-1', type: 'task.updated', actor_id: 999, ...overrides };
}

describe('useTaskRealtime', () => {
  beforeEach(() => {
    handler = undefined;
    vi.mocked(getCurrentUserId).mockReturnValue('42');
  });

  // A teammate reassigning this task away from the current user has to drop it
  // out of My tasks, not just refresh the open detail screen.
  it.each(['task.created', 'task.updated', 'task.status_changed', 'task.deleted'])(
    'invalidates my-tasks on %s',
    (type) => {
      const spy = mount();
      handler?.(event({ type }));
      expect(keys(spy)).toContain(MY_TASKS);
    }
  );

  it('still invalidates the task-scoped keys alongside my-tasks', () => {
    const spy = mount('t-9');
    handler?.(event({ topic: 'task:t-9' }));
    expect(keys(spy)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['task', 't-9']),
        JSON.stringify(['tasks']),
        MY_TASKS,
        JSON.stringify(['task-activity', 't-9']),
      ])
    );
  });

  it('ignores a task event the current user caused', () => {
    const spy = mount();
    handler?.(event({ actor_id: 42 }));
    expect(spy).not.toHaveBeenCalled();
  });

  // Comments never change assignment, so my-tasks stays untouched there.
  it('does not invalidate my-tasks for comment events', () => {
    const spy = mount();
    handler?.(event({ type: 'comment.created' }));
    expect(keys(spy)).not.toContain(MY_TASKS);
  });
});
