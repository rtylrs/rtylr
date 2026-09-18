import { createOneKitClient } from '@voxire/api';
import { TerminalServiceClient } from '@voxire/voxie/pos/terminal/v1/client';
import { useEffect } from 'react';
import { useTerminalContextStore } from '../../stores/terminalContextStore';
import { startTerminalHeartbeat } from './terminalHeartbeat';

const terminalClient = createOneKitClient(TerminalServiceClient, 'pos');

export function useTerminalHeartbeat(active: boolean): void {
  const terminalId = useTerminalContextStore((s) => s.terminalId);

  useEffect(() => {
    if (!active || !terminalId) return;
    return startTerminalHeartbeat({
      heartbeat: async () => {
        await terminalClient.heartbeatTerminal({ terminalId: String(terminalId) });
      },
    });
  }, [active, terminalId]);
}
