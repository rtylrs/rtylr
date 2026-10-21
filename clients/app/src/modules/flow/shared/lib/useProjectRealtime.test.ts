// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { RealtimeEvent } from '@voxire/api';

// Capture the handler the hook registers so events can be pushed through it
// directly - no WebSocket, no server. Mocking './realtime' also keeps the real
// module (which opens a client at import time) out of the test run.
let handler: ((e: RealtimeEvent) => void) | undefined;
vi.mock('./realtime', () => ({
  useRealtime: (_topics: string[], h: (e: RealtimeEvent) => void) => {
    handler = h;
  },
}));

vi.mock('./currentUser', () => ({
  getCurrentUserId: vi.fn(() => '42'),
}));

import { useProjectRealtime } from './useProjectRealtime';
import { getCurrentUserId } from './currentUser';

function mount(projectId = 'p-1') {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const spy = vi.spyOn(client, 'invalidateQueries');
  renderHook(() => useProjectRealtime(projectId), {
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children),
  });
  return spy;
}

type InvalidateSpy = ReturnType<typeof mount>;

/** The query keys the handler asked to invalidate, as comparable JSON. */
function keys(spy: InvalidateSpy): string[] {
  return spy.mock.calls.map((call) => JSON.stringify(call[0]?.queryKey));
}

const MY_TASKS = JSON.stringify(['my-tasks']);

/** A teammate-authored event (actor 999 is not the current user, 42). */
function event(overrides: Partial<RealtimeEvent> = {}): RealtimeEvent {
  return { topic: 'project:p-1', type: 'task.updated', actor_id: 999, ...overrides };
}

describe('useProjectRealtime', () => {
  beforeEach(() => {
    handler = undefined;
    vi.mocked(getCurrentUserId).mockReturnValue('42');
  });

  it.each([
    'task.created',
    'task.updated',
    'task.status_changed',
    'task.deleted',
    'column.updated',
  ])('invalidates my-tasks on %s so My tasks and Today go stale', (type) => {
    const spy = mount();
    handler?.(event({ type }));
    expect(keys(spy)).toContain(MY_TASKS);
  });

  it('still invalidates the project-scoped keys alongside my-tasks', () => {
    const spy = mount('p-7');
    handler?.(event());
    expect(keys(spy)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['tasks']),
        MY_TASKS,
        JSON.stringify(['columns', 'p-7']),
        JSON.stringify(['project-overview', 'p-7']),
        JSON.stringify(['milestones', 'p-7']),
      ])
    );
  });

  it('invalidates status projections when a teammate edits a status', () => {
    const spy = mount('p-7');
    handler?.(event({ type: 'flowStatus.updated' }));

    expect(keys(spy)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['statuses', 'p-7']),
        JSON.stringify(['tasks']),
        MY_TASKS,
        JSON.stringify(['task']),
      ])
    );
  });

  // The mutation's own onSuccess already invalidated; refetching again would
  // clobber an in-flight optimistic update.
  it('ignores a task event the current user caused', () => {
    const spy = mount();
    handler?.(event({ actor_id: 42 }));
    expect(spy).not.toHaveBeenCalled();
  });

  it('does not invalidate my-tasks for project-only events', () => {
    const spy = mount();
    handler?.(event({ type: 'project.updated' }));
    expect(keys(spy)).not.toContain(MY_TASKS);
  });
});
