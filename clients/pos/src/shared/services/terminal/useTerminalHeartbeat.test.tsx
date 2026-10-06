import { cleanup, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useTerminalContextStore } from '../../stores/terminalContextStore';
import { useTerminalHeartbeat } from './useTerminalHeartbeat';

const terminalClient = vi.hoisted(() => ({
  heartbeatTerminal: vi.fn(() => Promise.resolve({ message: 'ok' })),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => terminalClient,
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

    expect(terminalClient.heartbeatTerminal).not.toHaveBeenCalled();
  });

  it('does not start heartbeat without a selected terminal', async () => {
    useTerminalContextStore.setState({ locationId: 1, terminalId: null });

    renderHook(() => useTerminalHeartbeat(true));
    await Promise.resolve();

    expect(terminalClient.heartbeatTerminal).not.toHaveBeenCalled();
  });

  it('sends heartbeat with the selected terminal id', async () => {
    useTerminalContextStore.getState().setTerminalContext(1, 9);

    renderHook(() => useTerminalHeartbeat(true));

    await waitFor(() => {
      expect(terminalClient.heartbeatTerminal).toHaveBeenCalledWith({ terminalId: '9' });
    });
  });

  it('clears a terminal context that the server no longer recognizes', async () => {
    useTerminalContextStore.getState().setTerminalContext(1, 9);
    terminalClient.heartbeatTerminal.mockRejectedValueOnce(
      Object.assign(new Error('Terminal not found'), { statusCode: 404 })
    );

    renderHook(() => useTerminalHeartbeat(true));

    await waitFor(() => {
      expect(useTerminalContextStore.getState().terminalId).toBeNull();
      expect(useTerminalContextStore.getState().locationId).toBeNull();
    });
  });
});
