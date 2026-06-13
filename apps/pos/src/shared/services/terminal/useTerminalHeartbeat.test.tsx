import { cleanup, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { posService } from '../api/posService';
import { useTerminalContextStore } from '../../stores/terminalContextStore';
import { useTerminalHeartbeat } from './useTerminalHeartbeat';

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
}));

vi.mock('../api/posService', () => ({
  posService: {
    heartbeat: vi.fn(() => Promise.resolve({ data: {} })),
  },
}));

describe('useTerminalHeartbeat', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useTerminalContextStore.setState({ locationId: null, terminalId: null });
  });

  afterEach(() => {
    cleanup();
    useTerminalContextStore.setState({ locationId: null, terminalId: null });
  });

  it('does not start heartbeat when inactive', async () => {
    useTerminalContextStore.getState().setTerminalContext(1, 9);

    renderHook(() => useTerminalHeartbeat(false));
    await Promise.resolve();

    expect(posService.heartbeat).not.toHaveBeenCalled();
  });

  it('does not start heartbeat without a selected terminal', async () => {
    useTerminalContextStore.setState({ locationId: 1, terminalId: null });

    renderHook(() => useTerminalHeartbeat(true));
    await Promise.resolve();

    expect(posService.heartbeat).not.toHaveBeenCalled();
  });

  it('sends heartbeat with the selected terminal id', async () => {
    useTerminalContextStore.getState().setTerminalContext(1, 9);

    renderHook(() => useTerminalHeartbeat(true));

    await waitFor(() => {
      expect(posService.heartbeat).toHaveBeenCalledWith(9);
    });
  });
});
