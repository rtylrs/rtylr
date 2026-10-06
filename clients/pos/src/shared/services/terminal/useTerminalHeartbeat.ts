import { createOneKitClient } from '@voxire/api';
import { TerminalServiceClient } from '@voxire/voxie/pos/terminal/v1/client';
import { useEffect } from 'react';
import { useTerminalContextStore } from '../../stores/terminalContextStore';
import { startTerminalHeartbeat } from './terminalHeartbeat';

const terminalClient = createOneKitClient(TerminalServiceClient, 'pos');

export function useTerminalHeartbeat(active: boolean): void {
  const terminalId = useTerminalContextStore((s) => s.terminalId);
  const clearTerminalContext = useTerminalContextStore((s) => s.clearTerminalContext);

  useEffect(() => {
    if (!active || !terminalId) return;
    return startTerminalHeartbeat({
      heartbeat: async () => {
        try {
          await terminalClient.heartbeatTerminal({ terminalId: String(terminalId) });
        } catch (error) {
          if (httpStatus(error) === 404) {
            clearTerminalContext();
            return;
          }
          throw error;
        }
      },
    });
  }, [active, clearTerminalContext, terminalId]);
}

function httpStatus(error: unknown): number | undefined {
  if (!error || typeof error !== 'object') return undefined;
  const value = error as { status?: unknown; statusCode?: unknown; response?: { status?: unknown } };
  const status = value.statusCode ?? value.status ?? value.response?.status;
  return typeof status === 'number' ? status : undefined;
}
