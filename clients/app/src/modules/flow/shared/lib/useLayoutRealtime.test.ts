// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { RealtimeEvent } from '@voxire/api';

// Same harness as useProjectRealtime.test.ts / useTaskRealtime.test.ts: capture
// the handler the hook registers so events can be pushed through it directly,
// and keep the real './realtime' module (which opens a client at import time)
// out of the test run.
let handler: ((e: RealtimeEvent) => void) | undefined;
let topics: string[] | undefined;
vi.mock('./realtime', () => ({
  useRealtime: (t: string[], h: (e: RealtimeEvent) => void) => {
    topics = t;
    handler = h;
  },
}));

vi.mock('./currentUser', () => ({
  getCurrentUserId: vi.fn(() => '42'),
}));

import { useLayoutRealtime } from './useLayoutRealtime';
import { getCurrentUserId } from './currentUser';

// businessId is NOT defaulted: a default parameter fires on an explicit
// `undefined`, so a caller trying to assert the no-id case would silently get
// the default back. That cost a confusing red test.
function mount(businessId: string | undefined, userId: string | undefined = '42', enabled = true) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const spy = vi.spyOn(client, 'invalidateQueries');
  renderHook(() => useLayoutRealtime(businessId, userId, enabled), {
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

/** A teammate-authored event (actor 999 is not the current user, 42). */
function event(overrides: Partial<RealtimeEvent> = {}): RealtimeEvent {
  return { topic: 'business:1', type: 'project.created', actor_id: 999, ...overrides };
}

describe('useLayoutRealtime', () => {
  beforeEach(() => {
    handler = undefined;
    topics = undefined;
    vi.mocked(getCurrentUserId).mockReturnValue('42');
  });

  describe('subscription topics', () => {
    it('subscribes to the business and the user topic', () => {
      mount('1', '42');
      expect(topics).toEqual(['business:1', 'user:42']);
    });

    // No workspace predicate any more: the topic is the business, so the only
    // question is whether an id resolved at all.
    it('does not subscribe without a resolved business id', () => {
      mount(undefined, '42');
      expect(topics).toEqual(['user:42']);
    });

    it('subscribes to nothing when disabled', () => {
      mount('1', '42', false);
      expect(topics).toEqual([]);
    });
  });

  describe('invalidation fanout', () => {
    it('invalidates projects on a project event', () => {
      const spy = mount('1');
      handler?.(event({ type: 'project.created' }));
      // archived-projects too: archive/restore/purge move a project between the
      // list and the recycle bin, so the bin's query has to refetch as well.
      expect(keys(spy)).toEqual([
        JSON.stringify(['projects']),
        JSON.stringify(['projects-v2']),
        JSON.stringify(['archived-projects']),
      ]);
    });

    it('invalidates notifications on a notification event', () => {
      const spy = mount('1');
      handler?.(event({ topic: 'user:42', type: 'notification.created' }));
      expect(keys(spy)).toEqual([JSON.stringify(['flow-notifications'])]);
    });
  });

  // The mutation's own onSuccess already invalidated; refetching again is
  // wasteful and clobbers an in-flight optimistic update. Matches
  // useTaskRealtime and useProjectRealtime.
  describe('self-authored events', () => {
    it.each(['project.created', 'project.updated'])(
      'ignores %s when the current user is the actor',
      (type) => {
        const spy = mount('1');
        handler?.(event({ type, actor_id: 42 }));
        expect(spy).not.toHaveBeenCalled();
      }
    );

    it('processes notification events even when the current user is the actor', () => {
      const spy = mount('1');
      handler?.(event({ type: 'notification.created', actor_id: 42 }));
      expect(keys(spy)).toEqual([JSON.stringify(['flow-notifications'])]);
    });

    // actor_id arrives numeric from the server while the store holds a string.
    it('compares actor_id and user id as strings', () => {
      vi.mocked(getCurrentUserId).mockReturnValue('7');
      const spy = mount('1', '7');
      handler?.(event({ actor_id: 7 }));
      expect(spy).not.toHaveBeenCalled();
    });

    it('still handles a teammate event when the actor is someone else', () => {
      const spy = mount('1');
      handler?.(event({ actor_id: 999 }));
      expect(spy).toHaveBeenCalled();
    });

    // Fail open: an event with no actor, or no known current user, must still
    // invalidate - dropping it would leave the sidebar stale forever.
    it('handles an event with no actor_id', () => {
      const spy = mount('1');
      handler?.(event({ actor_id: undefined }));
      expect(spy).toHaveBeenCalled();
    });

    it('handles events when the current user is unknown', () => {
      vi.mocked(getCurrentUserId).mockReturnValue(null);
      const spy = mount('1');
      handler?.(event({ actor_id: 42 }));
      expect(spy).toHaveBeenCalled();
    });
  });
});
