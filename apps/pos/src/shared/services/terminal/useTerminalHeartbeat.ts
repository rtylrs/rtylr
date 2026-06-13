import { unwrap } from '@voxire/api';
import { useEffect } from 'react';
import { posService } from '../api/posService';
import { useTerminalContextStore } from '../../stores/terminalContextStore';
import { startTerminalHeartbeat } from './terminalHeartbeat';

export function useTerminalHeartbeat(active: boolean): void {
  const terminalId = useTerminalContextStore((s) => s.terminalId);

  useEffect(() => {
    if (!active || !terminalId) return;
    return startTerminalHeartbeat({
      heartbeat: async () => {
        await unwrap(posService.heartbeat(terminalId));
      },
    });
  }, [active, terminalId]);
}
