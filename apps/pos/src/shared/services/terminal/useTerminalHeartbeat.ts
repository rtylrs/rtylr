import { unwrap } from '@voxire/api';
import { useEffect } from 'react';
import { posService } from '../api/posService';
import { startTerminalHeartbeat } from './terminalHeartbeat';

export function useTerminalHeartbeat(active: boolean): void {
  useEffect(() => {
    if (!active) return;
    return startTerminalHeartbeat({
      heartbeat: async () => {
        await unwrap(posService.heartbeat());
      },
    });
  }, [active]);
}
