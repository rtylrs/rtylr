import { describe, it, expect, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useKdsStream } from './useKdsStream';

const realtime = vi.hoisted(() => ({ subscribe: vi.fn(() => vi.fn()) }));

vi.mock('../../shared/services/api/realtime', () => ({ posRealtime: realtime }));

describe('useKdsStream', () => {
  it('subscribes to the business topic, which every member may read', () => {
    renderHook(() => useKdsStream(42, () => {}));
    expect(realtime.subscribe).toHaveBeenCalledWith('business:42', expect.any(Function));
  });

  it('waits for the business ID', () => {
    realtime.subscribe.mockClear();
    renderHook(() => useKdsStream(undefined, () => {}));
    expect(realtime.subscribe).not.toHaveBeenCalled();
  });
});
