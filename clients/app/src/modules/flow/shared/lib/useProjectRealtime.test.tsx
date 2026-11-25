import React from 'react';
import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi } from 'vitest';
import type { RealtimeEvent } from '@voxire/api';

const handlers = vi.hoisted(() => [] as ((e: RealtimeEvent) => void)[]);

vi.mock('./realtime', () => ({
  useRealtime: (_topics: string[], handler: (e: RealtimeEvent) => void) => {
    handlers.push(handler);
  },
}));
vi.mock('./currentUser', () => ({ getCurrentUserId: () => '1' }));

import { useProjectRealtime } from './useProjectRealtime';

describe('useProjectRealtime', () => {
  it("refreshes an open cycle page and project progress when a teammate's task changes", () => {
    const qc = new QueryClient();
    const invalidate = vi.spyOn(qc, 'invalidateQueries');
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={qc}>{children}</QueryClientProvider>
    );
    renderHook(() => useProjectRealtime('3'), { wrapper });

    handlers.at(-1)?.({ type: 'task.updated', actor_id: 2 } as unknown as RealtimeEvent);

    const keys = invalidate.mock.calls.map(([filters]) => filters?.queryKey);
    expect(keys).toContainEqual(['cycles', 'detail']);
    expect(keys).toContainEqual(['projects-v2']);
    expect(keys).toContainEqual(['insights-dashboard-project', '3']);
  });
});
